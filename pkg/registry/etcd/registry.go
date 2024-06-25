package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/registry"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"log/slog"
	"sync"
)

var typesMap = map[mvccpb.Event_EventType]registry.EventType{
	mvccpb.PUT:    registry.EventTypeAdd,
	mvccpb.DELETE: registry.EventTypeDelete,
}

type Registry struct {
	sess   *concurrency.Session
	client *clientv3.Client

	mutex       sync.RWMutex
	watchCancel []func()
}

func NewRegistry(c *clientv3.Client) (*Registry, error) {
	sess, err := concurrency.NewSession(c)
	if err != nil {
		return nil, err
	}

	return &Registry{
		sess:   sess,
		client: c,
	}, nil
}

func (r *Registry) Register(ctx context.Context, si registry.Instance) error {
	val, err := json.Marshal(si)
	if err != nil {
		return nil
	}

	_, err = r.client.Put(ctx, r.instanceKey(si),
		string(val), clientv3.WithLease(r.sess.Lease()))

	return err
}

func (r *Registry) UnRegister(ctx context.Context, si registry.Instance) error {
	_, err := r.client.Delete(ctx, r.instanceKey(si))
	return err
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
	ctx, cancel := context.WithCancel(context.Background())
	ctx = clientv3.WithRequireLeader(ctx)
	r.mutex.Lock()
	r.watchCancel = append(r.watchCancel, cancel)
	r.mutex.Unlock()
	ch := r.client.Watch(ctx, r.serviceKey(name), clientv3.WithPrefix())
	res := make(chan registry.Event)
	go func() {
		for {
			select {
			case resp := <-ch:
				if resp.Canceled {
					return
				}
				if resp.Err() != nil {
					continue
				}

				for _, event := range resp.Events {
					var instance registry.Instance
					if event.Type == mvccpb.PUT {
						if err := json.Unmarshal(event.Kv.Value, &instance); err != nil {
							slog.Error("解析失败",
								slog.Any("name", "watch task"),
								slog.Any("error", err),
							)
						}
					}
					re := registry.Event{
						Type:     typesMap[event.Type],
						Key:      string(event.Kv.Key),
						Instance: instance,
					}
					res <- re
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return res
}

func (r *Registry) Close() error {
	r.mutex.Lock()
	for _, cancel := range r.watchCancel {
		cancel()
	}
	r.mutex.Unlock()
	return r.sess.Close()
}
