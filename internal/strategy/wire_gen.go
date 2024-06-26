// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package strategy

import (
	"github.com/Duke1616/ecmdb/internal/strategy/internal/service"
	"github.com/Duke1616/ecmdb/internal/strategy/internal/web"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/google/wire"
)

// Injectors from wire.go:

func InitModule(templateModule *template.Module) (*Module, error) {
	serviceService := service.NewService()
	service2 := templateModule.Svc
	handler := web.NewHandler(serviceService, service2)
	module := &Module{
		Hdl: handler,
	}
	return module, nil
}

// wire.go:

var ProviderSet = wire.NewSet(web.NewHandler, service.NewService)
