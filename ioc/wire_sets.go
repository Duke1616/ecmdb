package ioc

import (
	"github.com/Duke1616/ecmdb/internal/repository"
	"github.com/Duke1616/ecmdb/internal/repository/dao"
	attrSvc "github.com/Duke1616/ecmdb/internal/service/attribute"
	dataioSvc "github.com/Duke1616/ecmdb/internal/service/dataio"
	modelSvc "github.com/Duke1616/ecmdb/internal/service/model"
	pluginSvc "github.com/Duke1616/ecmdb/internal/service/plugin"
	relationSvc "github.com/Duke1616/ecmdb/internal/service/relation"
	resourceSvc "github.com/Duke1616/ecmdb/internal/service/resource"
	toolsSvc "github.com/Duke1616/ecmdb/internal/service/tools"
	attribute "github.com/Duke1616/ecmdb/internal/web/attribute"
	dataio "github.com/Duke1616/ecmdb/internal/web/dataio"
	model "github.com/Duke1616/ecmdb/internal/web/model"
	plugin "github.com/Duke1616/ecmdb/internal/web/plugin"
	relation "github.com/Duke1616/ecmdb/internal/web/relation"
	resource "github.com/Duke1616/ecmdb/internal/web/resource"
	terminal "github.com/Duke1616/ecmdb/internal/web/terminal"
	tools "github.com/Duke1616/ecmdb/internal/web/tools"
	"github.com/Duke1616/ecmdb/pkg/storage"
	"github.com/google/wire"
)

var (
	// BaseSet 基础设施 Provider 集合
	BaseSet = wire.NewSet(
		InitListener,
		InitMongoDB,
		InitMongoDBV2,
		InitEtcdClient,
		InitMQ,
		InitCrypto,
		InitMinioClient,
		storage.NewS3Storage,
	)

	AttributeSet = wire.NewSet(
		dao.NewAttributeDAO,
		dao.NewAttributeGroupDAO,
		repository.NewAttributeRepository,
		repository.NewAttributeGroupRepository,
		attrSvc.NewService,
		attribute.NewHandler,
		InitFieldSecureAttrChangeEventProducer,
		InitFieldDeleteEventProducer,
	)

	toolsSet = wire.NewSet(
		toolsSvc.NewService,
		tools.NewHandler,
	)

	terminalSet = wire.NewSet(
		terminal.NewHandler,
	)

	dataIoSet = wire.NewSet(
		dataio.NewHandler,
		dataioSvc.NewService,
	)

	PluginSet = wire.NewSet(
		dao.NewPluginDAO,
		repository.NewPluginRepository,
		pluginSvc.NewService,
		plugin.NewHandler,
	)

	RelationSet = wire.NewSet(
		dao.NewRelationTypeDAO,
		dao.NewRelationModelDAO,
		dao.NewRelationResourceDAO,
		repository.NewRelationTypeRepository,
		repository.NewRelationModelRepository,
		repository.NewRelationResourceRepository,
		relationSvc.NewRelationTypeService,
		relationSvc.NewRelationModelService,
		relationSvc.NewRelationResourceService,
		relation.NewRelationTypeHandler,
	)

	ModelSet = wire.NewSet(
		dao.NewModelDAO,
		dao.NewModelGroupDAO,
		repository.NewModelRepository,
		repository.NewModelGroupRepository,
		modelSvc.NewModelService,
		modelSvc.NewMGService,
		model.NewHandler,
	)

	ResourceSet = wire.NewSet(
		dao.NewResourceDAO,
		repository.NewResourceRepository,
		resourceSvc.NewService,
		resource.NewHandler,
	)

	// WebSet Web 服务 Provider 集合
	WebSet = wire.NewSet(
		InitPolicySDK,
		InitPermSyncer,
		InitProviders,
		InitGinMiddlewares,
		InitWebServer,

		BaseSet,
		AttributeSet,
		toolsSet,
		terminalSet,
		dataIoSet,
		PluginSet,
		RelationSet,
		ModelSet,
		ResourceSet,

		InitFieldSecureAttrChangeConsumer,
		InitFieldDeleteConsumer,
		InitTasks,

		InitDeleteModelDependencyCheckers,
		wire.Bind(new(modelSvc.IDefaultAttributeCreator), new(attrSvc.Service)),
	)
)

func InitDeleteModelDependencyCheckers(
	resourceSvc resourceSvc.Service,
	relationRMSvc relationSvc.RelationModelService,
) []modelSvc.IDeleteModelDependencyChecker {
	return []modelSvc.IDeleteModelDependencyChecker{
		resourceSvc,
		relationRMSvc,
	}
}
