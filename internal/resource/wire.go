//go:build wireinject

package resource

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/resource/internal/service"
	"github.com/Duke1616/ecmdb/internal/resource/internal/web"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	repository.NewResourceRepository,
	dao.NewResourceDAO)

func InitModule(db *mongo.Client, attributeModule *attribute.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		NewService,
		wire.FieldsOf(new(*attribute.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func NewService(repo repository.ResourceRepository) Service {
	return service.NewService(repo)
}
