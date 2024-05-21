// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/internal/relation/internal/web"
)

// Injectors from wire.go:

func InitRMHandler() (*web.RelationModelHandler, error) {
	mongo := InitMongoDB()
	module, err := relation.InitModule(mongo)
	if err != nil {
		return nil, err
	}
	relationModelHandler := module.RMHdl
	return relationModelHandler, nil
}

func InitRRHandler() (*web.RelationResourceHandler, error) {
	mongo := InitMongoDB()
	module, err := relation.InitModule(mongo)
	if err != nil {
		return nil, err
	}
	relationResourceHandler := module.RRHdl
	return relationResourceHandler, nil
}

func InitRRSvc() service.RelationResourceService {
	mongo := InitMongoDB()
	relationResourceService := relation.InitRRService(mongo)
	return relationResourceService
}

func InitRMSvc() service.RelationModelService {
	mongo := InitMongoDB()
	relationModelService := relation.InitRMService(mongo)
	return relationModelService
}