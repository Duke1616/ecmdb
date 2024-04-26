//go:build wireinject

package startup

import (
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/google/wire"
)

func InitHandler() (*model.Handler, error) {
	wire.Build(InitMongoDB,
		model.InitModule,
		wire.FieldsOf(new(*model.Module), "Hdl"),
	)
	return new(model.Handler), nil
}
