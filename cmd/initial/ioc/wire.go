//go:build wireinject

package ioc

import (
	"github.com/Duke1616/ecmdb/cmd/initial/version"
	bootstrapSvc "github.com/Duke1616/ecmdb/internal/service/bootstrap"
	attrSvc "github.com/Duke1616/ecmdb/internal/service/attribute"
	modelSvc "github.com/Duke1616/ecmdb/internal/service/model"
	"github.com/Duke1616/ecmdb/ioc"
	"github.com/google/wire"
)

var BaseSet = wire.NewSet(
	ioc.InitMongoDB,
	ioc.InitMongoDBV2,
	ioc.InitMySQLDB,
	ioc.InitRedis,
	ioc.InitMQ,
	ioc.InitEtcdClient,
	ioc.InitCrypto,
)

func InitApp() (*App, error) {
	wire.Build(
		wire.Struct(new(App), "*"),
		BaseSet,
		version.NewService,
		version.NewDao,
		bootstrapSvc.NewLoader,
		ioc.AttributeSet,
		ioc.RelationSet,
		ioc.ModelSet,
		ioc.ResourceSet,
		ioc.InitDeleteModelDependencyCheckers,
		wire.Bind(new(modelSvc.IDefaultAttributeCreator), new(attrSvc.Service)),
	)
	return new(App), nil
}
