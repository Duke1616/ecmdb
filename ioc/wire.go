//go:build wireinject

package ioc

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/google/wire"
)

var BaseSet = wire.NewSet(InitMongoDB)

func InitApp() (*App, error) {
	wire.Build(wire.Struct(new(App), "*"),
		BaseSet,
		model.InitHandler,
		attribute.InitHandler,
		//resource.InitHandler,
		InitWebServer,
		InitGinMiddlewares)
	return new(App), nil
}
