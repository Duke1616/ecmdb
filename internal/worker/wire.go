//go:build wireinject

package worker

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/worker/internal/event"
	"github.com/Duke1616/ecmdb/internal/worker/internal/event/watch"
	"github.com/Duke1616/ecmdb/internal/worker/internal/repository"
	"github.com/Duke1616/ecmdb/internal/worker/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/worker/internal/service"
	"github.com/Duke1616/ecmdb/internal/worker/internal/web"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewWorkerRepository,
)

func InitModule(q mq.MQ, db *mongox.Mongo, etcdClient *clientv3.Client) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitWorkerDAO,
		event.NewTaskRunnerEventProducer,
		initWatch,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func initWatch(etcdClient *clientv3.Client, svc service.Service) *watch.TaskWorkerWatch {
	task, err := watch.NewTaskWorkerWatch(etcdClient, svc)
	if err != nil {
		panic(err)
	}

	go task.Watch()
	return task
}

var (
	daoOnce = sync.Once{}
	d       dao.WorkerDAO
)

func InitProducer(producer event.TaskWorkerEventProducer) {
	wt, err := d.ListWorkerTopic(context.Background())
	if err != nil {
		panic(err)
	}

	// 开启 producer
	for _, item := range wt {
		err = producer.AddProducer(item.Topic)
		if err != nil {
			panic(err)
		}
	}
}

func InitWorkerDAO(db *mongox.Mongo, producer event.TaskWorkerEventProducer) dao.WorkerDAO {
	daoOnce.Do(func() {
		d = dao.NewWorkerDAO(db)
		InitProducer(producer)
	})

	return d
}
