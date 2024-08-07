//go:build wireinject

package endpoint

import (
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/repository"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/service"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewEndpointRepository,
	dao.NewEndpointDAO,
)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
