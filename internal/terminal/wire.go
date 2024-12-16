//go:build wireinject

package terminal

import (
	"github.com/Duke1616/ecmdb/internal/terminal/internal/web"
	"github.com/google/wire"
)

func InitModule() (*web.Handler, error) {
	wire.Build(
		web.NewHandler,
	)
	return new(web.Handler), nil
}
