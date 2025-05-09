//go:build wireinject

package startup

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/internal/test/ioc"
	"github.com/google/wire"
)

func InitHandler(rmModule *relation.Module, attrModule *attribute.Module, resourceModule *resource.Module) (*model.Handler, error) {
	wire.Build(ioc.InitMongoDB,
		model.InitModule,
		wire.FieldsOf(new(*model.Module), "Hdl"),
	)
	return new(model.Handler), nil
}
