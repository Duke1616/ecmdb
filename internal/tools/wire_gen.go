// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package tools

import (
	"github.com/Duke1616/ecmdb/internal/tools/service"
	"github.com/Duke1616/ecmdb/internal/tools/web"
	"github.com/minio/minio-go/v7"
)

// Injectors from wire.go:

func InitModule(minioClient *minio.Client) (*web.Handler, error) {
	serviceService := service.NewService(minioClient)
	handler := web.NewHandler(serviceService)
	return handler, nil
}
