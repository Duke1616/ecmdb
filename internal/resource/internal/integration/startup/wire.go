//go:build wireinject

package startup

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/google/wire"
)

func InitHandler(am *attribute.Module) (*resource.Handler, error) {
	wire.Build(InitMongoDB,
		resource.InitModule,
		wire.FieldsOf(new(*resource.Module), "Hdl"),
	)
	return new(resource.Handler), nil
}
