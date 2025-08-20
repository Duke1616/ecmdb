package watch

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Duke1616/ecmdb/internal/worker/internal/domain"
	"github.com/Duke1616/ecmdb/internal/worker/internal/service"
	"github.com/Duke1616/ecmdb/pkg/registry"
	"github.com/Duke1616/ecmdb/pkg/registry/etcd"
	"github.com/ecodeclub/ekit/slice"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const SUBSCRIBE = "worker"

type TaskWorkerWatch struct {
	svc         service.Service
	registry    registry.Registry
	svcWorkers  []domain.Worker
	etcdWorkers []registry.Instance
	close       chan struct{}
}

func NewTaskWorkerWatch(etcdClient *clientv3.Client, svc service.Service) (*TaskWorkerWatch, error) {
	r, err := etcd.NewRegistry(etcdClient)
	if err != nil {
		return nil, err
	}

	return &TaskWorkerWatch{
		svc:      svc,
		registry: r,
		close:    make(chan struct{}, 1),
	}, nil

}

func (t *TaskWorkerWatch) Watch() {
	err := t.compensation()
	if err != nil {
		slog.Error("补偿机制失败",
			slog.Any("error", err),
		)
	}

	// 监听数据
	events := t.registry.Subscribe(SUBSCRIBE)
	for {
		select {
		case event := <-events:
			t.handleEvent(event)
		case <-t.close:
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (t *TaskWorkerWatch) Close() {
	t.close <- struct{}{}
}

// compensation 启动补偿机制
func (t *TaskWorkerWatch) compensation() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var (
		err error
	)

	t.svcWorkers, _, err = t.svc.ListWorker(ctx, 0, 100)
	if err != nil {
		return fmt.Errorf("无法获取数据库的workers: %w", err)
	}

	t.etcdWorkers, err = t.registry.ListWorkers(ctx, SUBSCRIBE)
	if err != nil {
		return fmt.Errorf("无法获取服务中心的workers: %w", err)
	}

	err = t.forward(ctx)
	if err != nil {
		return fmt.Errorf("正向补偿失败: %w", err)
	}

	return t.reverse(ctx)
}

// forward 正向补偿 ETCD存在，但是数据库不存在
func (t *TaskWorkerWatch) forward(ctx context.Context) error {
	m := slice.ToMap(t.svcWorkers, func(element domain.Worker) string {
		return element.Name
	})

	for _, item := range t.etcdWorkers {
		_, exists := m[item.Name]
		if !exists {
			worker := domain.Worker{
				Key:    fmt.Sprintf("/task/worker/%s", item.Name),
				Name:   item.Name,
				Desc:   item.Desc,
				Topic:  item.Topic,
				Status: domain.RUNNING,
			}
			_, err := t.svc.Register(ctx, worker)
			if err != nil {
				return fmt.Errorf("补偿注册失败: %w", err)
			}
		}
	}

	return nil
}

// reverse 反向补偿、如果ETCD中不存在、但是MongoDB中存在切状态为运行、需要修改为离线
func (t *TaskWorkerWatch) reverse(ctx context.Context) error {
	m := slice.ToMap(t.etcdWorkers, func(element registry.Instance) string {
		return element.Name
	})

	for _, item := range t.svcWorkers {
		_, exists := m[item.Name]
		if !exists {
			_, err := t.svc.UpdateStatus(ctx, item.Id, domain.OFFLINE.ToUint8())
			if err != nil {
				return fmt.Errorf("补偿修改状态失败: %w", err)
			}
		}
	}

	return nil
}

func (t *TaskWorkerWatch) handleEvent(event registry.Event) {
	name, err := t.svc.FindOrRegisterByKey(context.Background(), t.toDomain(event))
	if err != nil {
		slog.Error("新增 OR 修改工作节点失败",
			slog.Any("error", err),
			slog.Any("name", name),
		)
	}
}

func (t *TaskWorkerWatch) toDomain(event registry.Event) domain.Worker {
	worker := domain.Worker{
		Name:  event.Instance.Name,
		Desc:  event.Instance.Desc,
		Topic: event.Instance.Topic,
		Key:   event.Key,
	}

	switch event.Type {
	case registry.EventTypeAdd:
		worker.Status = domain.RUNNING
	case registry.EventTypeDelete:
		worker.Status = domain.OFFLINE
	}

	return worker
}
