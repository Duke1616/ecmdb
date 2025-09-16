package ioc

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/resource"
)

type App struct {
	ModelSvc    model.Service
	AttrSvc     attribute.Service
	ResourceSvc resource.Service
	AesKey      string
}
