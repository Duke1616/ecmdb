//go:build wireinject

package ioc

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/google/wire"
)

var BaseSet = wire.NewSet(InitMongoDB)

func InitApp() (*App, error) {
	wire.Build(wire.Struct(new(App), "*"),
		BaseSet,
		InitLdapConfig,
		model.InitModule,
		wire.FieldsOf(new(*model.Module), "Hdl"),
		attribute.InitModule,
		wire.FieldsOf(new(*attribute.Module), "Hdl"),
		resource.InitModule,
		wire.FieldsOf(new(*resource.Module), "Hdl"),
		relation.InitModule,
		wire.FieldsOf(new(*relation.Module), "RRHdl", "RMHdl", "RTHdl"),
		user.InitModule,
		wire.FieldsOf(new(*user.Module), "Hdl"),
		InitWebServer,
		InitGinMiddlewares)
	return new(App), nil
}
