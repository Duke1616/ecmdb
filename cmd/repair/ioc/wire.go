//go:build wireinject

package ioc

import (
	attrSvc "github.com/Duke1616/ecmdb/internal/service/attribute"
	modelSvc "github.com/Duke1616/ecmdb/internal/service/model"
	"github.com/Duke1616/ecmdb/ioc"
	"github.com/google/wire"
)

func InitApp() (*App, error) {
	wire.Build(
		wire.Struct(new(App), "*"),
		ioc.BaseSet,
		ioc.AttributeSet,
		ioc.RelationSet,
		ioc.ModelSet,
		ioc.ResourceSet,
		ioc.InitDeleteModelDependencyCheckers,
		wire.Bind(new(modelSvc.IDefaultAttributeCreator), new(attrSvc.Service)),
	)
	return new(App), nil
}
