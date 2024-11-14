//go:build wireinject

package rota

import (
	"github.com/Duke1616/ecmdb/internal/rota/internal/repository"
	"github.com/Duke1616/ecmdb/internal/rota/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/rota/internal/service"
	"github.com/Duke1616/ecmdb/internal/rota/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewRotaRepository,
	dao.NewRotaDao,
)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
