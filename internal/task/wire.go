package task

import (
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/google/wire"
)

func InitModule(orderModule *order.Module) (*Module, error) {
	wire.Build(
		service.NewService,
		wire.FieldsOf(new(*order.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
