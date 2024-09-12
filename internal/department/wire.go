//go:build wireinject

package department

import (
	"github.com/Duke1616/ecmdb/internal/department/internal/repository"
	"github.com/Duke1616/ecmdb/internal/department/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/department/internal/service"
	"github.com/Duke1616/ecmdb/internal/department/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewDepartmentRepository,
	dao.NewDepartmentDAO,
)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
