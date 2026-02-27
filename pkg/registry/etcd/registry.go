package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Duke1616/ecmdb/pkg/registry"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var typesMap = map[mvccpb.Event_EventType]registry.EventType{
	mvccpb.PUT:    registry.EventTypeAdd,
	mvccpb.DELETE: registry.EventTypeDelete,
}

type Registry struct {
	client *clientv3.Client

	mutex       sync.RWMutex
	sess        *concurrency.Session
	watchCancel []func()

	// registrations 记录所有已成功的注册实例。
	// 当 etcd Session 因为网络抖动租约到期后，monitorSession 会利用此 map 自动进行重新注册。
	registrations map[string]registry.Instance

	closeCtx context.Context
	cancel   context.CancelFunc
}

func NewRegistry(c *clientv3.Client) (*Registry, error) {
	sess, err := concurrency.NewSession(c)
	if err != nil {
		return nil, fmt.Errorf("创建 etcd 会话失败: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	r := &Registry{
		client:        c,
		sess:          sess,
		registrations: make(map[string]registry.Instance),
		closeCtx:      ctx,
		cancel:        cancel,
	}

	// 启动后台监控协程，处理 Session 节点的自动维护
	go r.monitorSession()

	return r, nil
}

func (r *Registry) Register(ctx context.Context, si registry.Instance) error {
	val, err := json.Marshal(si)
	if err != nil {
		return fmt.Errorf("实例序列化失败: %w", err)
	}

	key := r.instanceKey(si)

	r.mutex.Lock()
	r.registrations[key] = si
	sess := r.sess
	r.mutex.Unlock()

	// 写入 etcd，绑定当前 Lease
	_, err = r.client.Put(ctx, key, string(val), clientv3.WithLease(sess.Lease()))
	if err != nil {
		return fmt.Errorf("写入注册信息失败: %w", err)
	}

	return nil
}

func (r *Registry) UnRegister(ctx context.Context, si registry.Instance) error {
	key := r.instanceKey(si)

	r.mutex.Lock()
	delete(r.registrations, key)
	r.mutex.Unlock()

	_, err := r.client.Delete(ctx, key)
	return err
}

func (r *Registry) monitorSession() {
	for {
		r.mutex.RLock()
		sess := r.sess
		r.mutex.RUnlock()

		// 等待当前 Session 终止（租约到期或网络彻底断开）
		select {
		case <-sess.Done():
			slog.Warn("etcd 会话租约已过期或连接断开，尝试重建会话并重新注册")
		case <-r.closeCtx.Done():
			return
		}

		// 指数退避重建 Session
		newSess := r.retryNewSession()
		if newSess == nil {
			return // 应该是被 Close 了
		}

		r.mutex.Lock()
		r.sess = newSess
		// 自动恢复之前所有的注册节点
		for key, si := range r.registrations {
			r.reRegister(key, si)
		}
		r.mutex.Unlock()
	}
}

func (r *Registry) retryNewSession() *concurrency.Session {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-r.closeCtx.Done():
			return nil
		case <-ticker.C:
			sess, err := concurrency.NewSession(r.client)
			if err == nil {
				return sess
			}
			slog.Error("重建 etcd 会话失败，持续重试中...", slog.Any("error", err))
		}
	}
}

func (r *Registry) reRegister(key string, si registry.Instance) {
	val, _ := json.Marshal(si)
	_, err := r.client.Put(r.closeCtx, key, string(val), clientv3.WithLease(r.sess.Lease()))
	if err != nil {
		slog.Error("会话恢复后重新注册失败", slog.String("key", key), slog.Any("error", err))
	} else {
		slog.Info("会话恢复后重新注册成功", slog.String("key", key))
	}
}

func (r *Registry) ListWorkers(ctx context.Context, name string) ([]registry.Instance, error) {
	resp, err := r.client.Get(ctx, r.serviceKey(name), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	res := make([]registry.Instance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var si registry.Instance
		err = json.Unmarshal(kv.Value, &si)
		if err != nil {
			return nil, err
		}
		res = append(res, si)
	}
	return res, nil
}

func (r *Registry) instanceKey(s registry.Instance) string {
	return fmt.Sprintf("/task/worker/%s", s.Name)
}

func (r *Registry) serviceKey(name string) string {
	return fmt.Sprintf("/task/%s", name)
}

func (r *Registry) Subscribe(name string) <-chan registry.Event {
	res := make(chan registry.Event)
	ctx, cancel := context.WithCancel(r.closeCtx)

	r.mutex.Lock()
	r.watchCancel = append(r.watchCancel, cancel)
	r.mutex.Unlock()

	go func() {
		defer close(res)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Watch 需要 RequireLeader 确保连上的是健康的集群
			watchCtx := clientv3.WithRequireLeader(ctx)
			ch := r.client.Watch(watchCtx, r.serviceKey(name), clientv3.WithPrefix())

			for resp := range ch {
				if resp.Canceled {
					slog.Warn("etcd watch 断开，准备重连",
						slog.String("service", name),
						slog.Any("error", resp.Err()))
					break
				}

				if err := resp.Err(); err != nil {
					slog.Error("etcd watch 错误",
						slog.String("service", name),
						slog.Any("error", err))
					break
				}

				for _, event := range resp.Events {
					var instance registry.Instance
					if event.Type == mvccpb.PUT {
						if err := json.Unmarshal(event.Kv.Value, &instance); err != nil {
							slog.Error("解析服务实例失败", slog.Any("error", err))
						}
					}
					re := registry.Event{
						Type:     typesMap[event.Type],
						Key:      string(event.Kv.Key),
						Instance: instance,
					}

					select {
					case res <- re:
					case <-ctx.Done():
						return
					}
				}
			}

			// 避免紧凑死循环
			select {
			case <-time.After(time.Second * 3):
			case <-ctx.Done():
				return
			}
		}
	}()

	return res
}

func (r *Registry) Close() error {
	r.cancel()
	r.mutex.Lock()
	for _, cancel := range r.watchCancel {
		cancel()
	}
	r.mutex.Unlock()
	return r.sess.Close()
}
