//go:build wireinject

package worker

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/runner"
	"github.com/Duke1616/ecmdb/internal/worker/internal/event"
	"github.com/Duke1616/ecmdb/internal/worker/internal/repository"
	"github.com/Duke1616/ecmdb/internal/worker/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/worker/internal/service"
	"github.com/Duke1616/ecmdb/internal/worker/internal/web"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewWorkerRepository,
	dao.NewWorkerDAO)

func InitModule(q mq.MQ, db *mongox.Mongo, runnerModule *runner.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		initConsumer,
		wire.FieldsOf(new(*runner.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func initConsumer(svc service.Service, q mq.MQ) *event.TaskWorkerConsumer {
	consumer, err := event.NewTaskWorkerConsumer(svc, q)
	if err != nil {
		panic(err)
	}

	consumer.Start(context.Background())
	return consumer
}
