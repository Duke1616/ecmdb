//go:build wireinject

package startup

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/google/wire"
)

func InitHandler(attributeModule *attribute.Module, relationModule *relation.Module) (*resource.Handler, error) {
	wire.Build(InitMongoDB,
		InitMQ,
		InitCryptoRegistry,
		resource.InitModule,
		wire.FieldsOf(new(*resource.Module), "Hdl"),
	)
	return new(resource.Handler), nil
}
