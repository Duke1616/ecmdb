//go:build wireinject

package ioc

import (
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
	)
	return new(App), nil
}
