//go:build wireinject

package ioc

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/ioc"
	"github.com/google/wire"
)

var BaseSet = wire.NewSet(ioc.InitMongoDB, ioc.InitMQ, ioc.InitModuleCrypto)

func InitApp() (*App, error) {
	wire.Build(wire.Struct(new(App), "*"),
		BaseSet,
		relation.InitModule,
		model.InitModule,
		wire.FieldsOf(new(*model.Module), "Svc"),
		attribute.InitModule,
		wire.FieldsOf(new(*attribute.Module), "Svc"),
		resource.InitModule,
		wire.FieldsOf(new(*resource.Module), "EncryptedSvc"),
	)
	return new(App), nil
}
