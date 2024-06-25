package event

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/worker/internal/domain"
	"github.com/Duke1616/ecmdb/internal/worker/internal/service"
	"github.com/Duke1616/ecmdb/pkg/registry"
	"github.com/Duke1616/ecmdb/pkg/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log/slog"
	"time"
)

type TaskWorkerWatch struct {
	svc      service.Service
	registry registry.Registry
	close    chan struct{}
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
	events := t.registry.Subscribe("worker")
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
