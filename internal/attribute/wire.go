//go:build wireinject

package attribute

import (
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/service"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
	"sync"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	repository.NewAttributeRepository,
	repository.NewAttributeGroupRepository,
	dao.NewAttributeGroupDAO)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		NewService,
		InitAttributeDAO,
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

func InitAttributeDAO(db *mongox.Mongo) dao.AttributeDAO {
	InitCollectionOnce(db)
	return dao.NewAttributeDAO(db)
}

func NewService(repo repository.AttributeRepository, repoGroup repository.AttributeGroupRepository) Service {
	return service.NewService(repo, repoGroup)
}
