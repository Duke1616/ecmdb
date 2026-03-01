//go:build wireinject

package worker

import (
	"github.com/Duke1616/ecmdb/internal/worker/internal/event"
	"github.com/Duke1616/ecmdb/internal/worker/internal/job"
	"github.com/Duke1616/ecmdb/internal/worker/internal/service"
	"github.com/Duke1616/ework-runner/pkg/grpc/registry"
	"github.com/Duke1616/ework-runner/pkg/grpc/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	service.NewService,
	InitRegistry,
	wire.Bind(new(registry.Registry), new(*etcd.Registry)),
)

func InitModule(q mq.MQ, etcdClient *clientv3.Client) (*Module, error) {
	wire.Build(
		ProviderSet,
		event.NewTaskRunnerEventProducer,
		initWatch,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func initWatch(r registry.Registry, svc service.Service) *job.ServiceDiscoveryJob {
	task, err := job.NewServiceDiscoveryJob(r, svc)
	if err != nil {
		panic(err)
	}

	go task.Watch()
	return task
}

func InitRegistry(etcdClient *clientv3.Client) (*etcd.Registry, error) {
	return etcd.NewRegistryWithPrefix(etcdClient, "/etask/kafka")
}
