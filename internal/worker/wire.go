//go:build wireinject

package worker

import (
	"github.com/Duke1616/ecmdb/internal/runner"
	"github.com/Duke1616/ecmdb/internal/worker/internal/event"
	"github.com/Duke1616/ecmdb/internal/worker/internal/repository"
	"github.com/Duke1616/ecmdb/internal/worker/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/worker/internal/service"
	"github.com/Duke1616/ecmdb/internal/worker/internal/web"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewWorkerRepository,
	dao.NewWorkerDAO)

func InitModule(q mq.MQ, db *mongox.Mongo, etcdClient *clientv3.Client, runnerModule *runner.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		initWatch,
		wire.FieldsOf(new(*runner.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func initWatch(etcdClient *clientv3.Client, svc service.Service) *event.TaskWorkerWatch {
	task, err := event.NewTaskWorkerWatch(etcdClient, svc)
	if err != nil {
		panic(err)
	}

	go task.Watch()
	return task
}
