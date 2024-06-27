//go:build wireinject

package strategy

import (
	"github.com/Duke1616/ecmdb/internal/strategy/internal/service"
	"github.com/Duke1616/ecmdb/internal/strategy/internal/web"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
)

func InitModule(templateModule *template.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.FieldsOf(new(*template.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
