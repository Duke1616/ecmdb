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
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewRoleRepository,
	dao.NewRoleDAO,
)

func InitModule(db *mongox.Mongo, menuModule *menu.Module, policyModule *policy.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
		wire.FieldsOf(new(*menu.Module), "Svc"),
		wire.FieldsOf(new(*policy.Module), "Svc"),
	)
	return new(Module), nil
}
