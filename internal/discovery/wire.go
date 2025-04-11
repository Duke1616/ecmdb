//go:build wireinject

package discovery

import (
	"github.com/Duke1616/ecmdb/internal/discovery/internal/repository"
	"github.com/Duke1616/ecmdb/internal/discovery/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/discovery/internal/service"
	"github.com/Duke1616/ecmdb/internal/discovery/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewDiscoveryRepository,
	dao.NewDiscoveryDAO,
)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
