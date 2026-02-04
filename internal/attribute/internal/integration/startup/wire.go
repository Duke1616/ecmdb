//go:build wireinject

package startup

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/google/wire"
)

func InitHandler() (*attribute.Handler, error) {
	wire.Build(
		InitMongoDB,
		InitMQ,
		attribute.InitModule,
		wire.FieldsOf(new(*attribute.Module), "Hdl"),
	)
	return new(attribute.Handler), nil
}
