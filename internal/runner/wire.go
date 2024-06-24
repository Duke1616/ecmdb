//go:build wireinject

package runner

import (
	"github.com/Duke1616/ecmdb/internal/runner/event"
	"github.com/Duke1616/ecmdb/internal/runner/service"
	"github.com/Duke1616/ecmdb/internal/runner/web"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
)

func InitModule(q mq.MQ) (*Module, error) {
	wire.Build(
		web.NewHandler,
		event.NewTaskRunnerEventProducer,
		service.NewService,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
