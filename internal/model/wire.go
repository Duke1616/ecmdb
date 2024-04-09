package model

import (
	"github.com/Duke1616/ecmdb/internal/model/internal/web"
	"github.com/google/wire"
)

type Handler = web.Handler

var ProviderSet = wire.NewSet(web.NewHandler)

func InitHandler() *Handler {
	wire.Build(ProviderSet)
	return new(Handler)
}
