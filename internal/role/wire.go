//go:build wireinject

package role

import (
	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/role/internal/repository"
	"github.com/Duke1616/ecmdb/internal/role/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/role/internal/service"
	"github.com/Duke1616/ecmdb/internal/role/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
	"sync"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewRoleRepository,
)

func InitModule(db *mongox.Mongo, menuModule *menu.Module, policyModule *policy.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitRoleDAO,
		wire.Struct(new(Module), "*"),
		wire.FieldsOf(new(*menu.Module), "Svc"),
		wire.FieldsOf(new(*policy.Module), "Svc"),
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
