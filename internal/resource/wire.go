//go:build wireinject

package resource

import (
	"github.com/Duke1616/ecmdb/internal/resource/internal/web"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler)

func InitHandler() *Handler {
	wire.Build(ProviderSet)
	return new(Handler)
}

type Handler = web.Handler
