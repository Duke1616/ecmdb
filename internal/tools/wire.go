//go:build wireinject

package tools

import (
	"github.com/Duke1616/ecmdb/internal/tools/service"
	"github.com/Duke1616/ecmdb/internal/tools/web"
	"github.com/google/wire"
	"github.com/minio/minio-go/v7"
)

func InitModule(minioClient *minio.Client) (*web.Handler, error) {
	wire.Build(
		web.NewHandler,
		service.NewService,
	)
	return new(web.Handler), nil
}
