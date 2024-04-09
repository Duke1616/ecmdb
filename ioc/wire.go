//go:build wireinject

package ioc

import (
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/google/wire"
)

func InitApp() (*App, error) {
	wire.Build(wire.Struct(new(App), "*"),
		model.InitHandler,
		InitWebServer,
		InitGinMiddlewares)
	return new(App), nil
}
