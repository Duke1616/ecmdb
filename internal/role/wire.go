//go:build wireinject

package role

import (
	"sync"

	"github.com/Duke1616/ecmdb/internal/role/internal/repository"
	"github.com/Duke1616/ecmdb/internal/role/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/role/internal/service"
	"github.com/Duke1616/ecmdb/internal/role/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewRoleRepository,
)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitRoleDAO,
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

func InitRoleDAO(db *mongox.Mongo) dao.RoleDAO {
	InitCollectionOnce(db)
	return dao.NewRoleDAO(db)
}
