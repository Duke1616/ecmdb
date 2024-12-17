//go:build wireinject

package terminal

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/internal/terminal/internal/web"
	"github.com/google/wire"
)

func InitModule(relationModule *relation.Module, resourceModule *resource.Module, attributeModule *attribute.Module) (*web.Handler, error) {
	wire.Build(
		web.NewHandler,
		wire.FieldsOf(new(*relation.Module), "RRSvc"),
		wire.FieldsOf(new(*resource.Module), "Svc"),
		wire.FieldsOf(new(*attribute.Module), "Svc"),
	)
	return new(web.Handler), nil
}
