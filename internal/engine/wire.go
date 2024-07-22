//go:build wireinject

package engine

import (
	"github.com/Duke1616/ecmdb/internal/engine/internal/repository"
	"github.com/Duke1616/ecmdb/internal/engine/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/engine/internal/service"
	"github.com/Duke1616/ecmdb/internal/engine/internal/web"
	"github.com/google/wire"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewProcessEngineRepository,
	dao.NewProcessEngineDAO,
)

func InitModule(db *gorm.DB) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
