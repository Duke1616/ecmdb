//go:build wireinject

package runner

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/runner/internal/event"
	"github.com/Duke1616/ecmdb/internal/runner/internal/repository"
	"github.com/Duke1616/ecmdb/internal/runner/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/runner/internal/service"
	"github.com/Duke1616/ecmdb/internal/runner/internal/web"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewRunnerRepository,
	dao.NewRunnerDAO,
)

func InitModule(db *mongox.Mongo, q mq.MQ, workerModule *worker.Module, codebookModule *codebook.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		initTaskRunnerConsumer,
		wire.FieldsOf(new(*worker.Module), "Svc"),
		wire.FieldsOf(new(*codebook.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func initTaskRunnerConsumer(svc service.Service, mq mq.MQ) *event.TaskRunnerConsumer {
	consumer, err := event.NewTaskRunnerConsumer(svc, mq)
	if err != nil {
		panic(err)
	}

	consumer.Start(context.Background())
	return consumer
}
