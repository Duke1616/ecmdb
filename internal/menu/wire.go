//go:build wireinject

package menu

import (
	"github.com/Duke1616/ecmdb/internal/menu/internal/repository"
	"github.com/Duke1616/ecmdb/internal/menu/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/menu/internal/service"
	"github.com/Duke1616/ecmdb/internal/menu/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewMenuRepository,
	dao.NewMenuDAO,
)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
