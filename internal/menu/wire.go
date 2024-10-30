//go:build wireinject

package menu

import (
	"github.com/Duke1616/ecmdb/internal/menu/internal/event"
	"github.com/Duke1616/ecmdb/internal/menu/internal/repository"
	"github.com/Duke1616/ecmdb/internal/menu/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/menu/internal/service"
	"github.com/Duke1616/ecmdb/internal/menu/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewMenuRepository,
	dao.NewMenuDAO,
)

func InitModule(q mq.MQ, db *mongox.Mongo) (*Module, error) {
	wire.Build(
		event.NewMenuChangeEventProducer,
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
