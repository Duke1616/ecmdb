package ioc

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/repository"
	"github.com/Duke1616/ecmdb/internal/repository/dao"
	attrSvc "github.com/Duke1616/ecmdb/internal/service/attribute"
	modelSvc "github.com/Duke1616/ecmdb/internal/service/model"
	pluginSvc "github.com/Duke1616/ecmdb/internal/service/plugin"
	relationSvc "github.com/Duke1616/ecmdb/internal/service/relation"
	resourceSvc "github.com/Duke1616/ecmdb/internal/service/resource"
	rootioc "github.com/Duke1616/ecmdb/ioc"
	"github.com/google/wire"
)

var (
	commandBaseSet = wire.NewSet(
		rootioc.InitMongoDB,
		rootioc.InitMongoDBV2,
		rootioc.InitCrypto,
	)

	commandAttributeSet = wire.NewSet(
		InitNoopFieldSecureAttrChangeEventProducer,
		InitNoopFieldDeleteEventProducer,
		dao.NewAttributeDAO,
		dao.NewAttributeGroupDAO,
		repository.NewAttributeRepository,
		repository.NewAttributeGroupRepository,
		attrSvc.NewService,
		wire.Bind(new(modelSvc.IDefaultAttributeCreator), new(attrSvc.Service)),
	)

	commandResourceSet = wire.NewSet(
		dao.NewResourceDAO,
		repository.NewResourceRepository,
		resourceSvc.NewService,
	)

	commandModelSet = wire.NewSet(
		dao.NewModelDAO,
		dao.NewModelGroupDAO,
		repository.NewModelRepository,
		repository.NewModelGroupRepository,
		modelSvc.NewModelService,
		modelSvc.NewMGService,
		InitPluginCommandDeleteModelDependencyCheckers,
	)

	commandRelationSet = wire.NewSet(
		dao.NewRelationTypeDAO,
		dao.NewRelationModelDAO,
		dao.NewRelationResourceDAO,
		repository.NewRelationTypeRepository,
		repository.NewRelationModelRepository,
		repository.NewRelationResourceRepository,
		relationSvc.NewRelationTypeService,
		relationSvc.NewRelationModelService,
		relationSvc.NewRelationResourceService,
	)

	commandPluginSet = wire.NewSet(
		dao.NewPluginDAO,
		repository.NewPluginRepository,
		pluginSvc.NewService,
	)

	PluginCommandSet = wire.NewSet(
		commandBaseSet,
		commandAttributeSet,
		commandResourceSet,
		commandModelSet,
		commandRelationSet,
		commandPluginSet,
	)
)

type noopFieldSecureAttrChangeEventProducer struct{}

func InitNoopFieldSecureAttrChangeEventProducer() attrSvc.FieldSecureAttrChangeEventProducer {
	return noopFieldSecureAttrChangeEventProducer{}
}

func (noopFieldSecureAttrChangeEventProducer) Produce(context.Context, domain.FieldSecureAttrChange) error {
	return nil
}

type noopFieldDeleteEventProducer struct{}

func InitNoopFieldDeleteEventProducer() attrSvc.IFieldDeleteEventProducer {
	return noopFieldDeleteEventProducer{}
}

func (noopFieldDeleteEventProducer) Produce(context.Context, domain.FieldDelete) error {
	return nil
}

func InitPluginCommandDeleteModelDependencyCheckers() []modelSvc.IDeleteModelDependencyChecker {
	return nil
}
