package ioc

import (
	attrSvc "github.com/Duke1616/ecmdb/internal/service/attribute"
	modelSvc "github.com/Duke1616/ecmdb/internal/service/model"
	resourceSvc "github.com/Duke1616/ecmdb/internal/service/resource"
)

type App struct {
	ModelSvc    modelSvc.Service
	AttrSvc     attrSvc.Service
	ResourceSvc resourceSvc.EncryptedSvc
}
