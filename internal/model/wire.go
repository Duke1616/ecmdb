//go:build wireinject

package model

import (
	"github.com/Duke1616/ecmdb/internal/model/internal/repository"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/model/internal/service"
	"github.com/Duke1616/ecmdb/internal/model/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewModelRepository,
	dao.NewModelDAO)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
