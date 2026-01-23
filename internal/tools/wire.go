//go:build wireinject

package tools

import (
	"github.com/Duke1616/ecmdb/internal/tools/service"
	"github.com/Duke1616/ecmdb/internal/tools/web"
	"github.com/Duke1616/ecmdb/pkg/storage"
	"github.com/google/wire"
)

func InitModule(storage *storage.S3Storage) (*web.Handler, error) {
	wire.Build(
		web.NewHandler,
		service.NewService,
	)
	return new(web.Handler), nil
}
