//go:build wireinject

package relation

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/internal/relation/internal/web"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
)

var ProviderSet = wire.NewSet(
	web.NewRelationResourceHandler,
	web.NewRelationModelHandler,
	service.NewRelationResourceService,
	service.NewRelationModelService,
	repository.NewRelationModelRepository,
	repository.NewRelationResourceRepository,
	dao.NewRelationModelDAO,
	dao.NewRelationResourceDAO)

func InitModule(db *mongo.Client, attributeModel *attribute.Module, resourceModel *resource.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.FieldsOf(new(*attribute.Module), "Svc"),
		wire.FieldsOf(new(*resource.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
