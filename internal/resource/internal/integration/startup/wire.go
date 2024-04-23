//go:build wireinject

package startup

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/google/wire"
)

func InitHandler() (*resource.Handler, error) {
	wire.Build(InitMongoDB,
		resource.InitModule,
		attribute.InitModule,
		wire.FieldsOf(new(*resource.Module), "Hdl"),
	)
	return new(resource.Handler), nil
}
