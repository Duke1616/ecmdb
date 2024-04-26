//go:build wireinject

package resource

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/resource/internal/service"
	"github.com/Duke1616/ecmdb/internal/resource/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
	"sync"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	repository.NewResourceRepository)

func InitModule(db *mongox.Mongo, attributeModule *attribute.Module, relationModule *relation.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		NewService,
		InitResourceDAO,
		wire.FieldsOf(new(*attribute.Module), "Svc"),
		wire.FieldsOf(new(*relation.Module), "RRSvc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var daoOnce = sync.Once{}

func InitCollectionOnce(db *mongox.Mongo) {
	daoOnce.Do(func() {
		err := dao.InitIndexes(db)
		if err != nil {
			panic(err)
		}
	})
}

func InitResourceDAO(db *mongox.Mongo) dao.ResourceDAO {
	InitCollectionOnce(db)
	return dao.NewResourceDAO(db)
}

func NewService(repo repository.ResourceRepository) Service {
	return service.NewService(repo)
}
