//go:build wireinject

package tools

import (
	"github.com/Duke1616/ecmdb/internal/tools/web"
	"github.com/google/wire"
)

func InitModule() (*web.Handler, error) {
	wire.Build(
		web.NewHandler,
	)
	return new(web.Handler), nil
}
