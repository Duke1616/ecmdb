//go:build wireinject

package relation

import (
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/internal/relation/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewRelationResourceHandler,
	web.NewRelationModelHandler,
	web.NewRelationTypeHandler,
	service.NewRelationResourceService,
	service.NewRelationModelService,
	service.NewRelationTypeService,
	repository.NewRelationModelRepository,
	repository.NewRelationResourceRepository,
	repository.NewRelationTypeRepository,
	dao.NewRelationModelDAO,
	dao.NewRelationResourceDAO,
	dao.NewRelationTypeDAO)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
