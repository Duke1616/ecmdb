//go:build wireinject

package attribute

import (
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/service"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/web"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	repository.NewAttributeRepository,
	dao.NewAttributeDAO)

func InitModule(db *mongo.Client) (*Module, error) {
	wire.Build(
		ProviderSet,
		NewService,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func NewService(repo repository.AttributeRepository) Service {
	return service.NewService(repo)
}
