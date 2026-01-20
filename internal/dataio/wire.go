//go:build wireinject

package dataio

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/dataio/internal/service"
	"github.com/Duke1616/ecmdb/internal/dataio/internal/web"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/pkg/storage"
	"github.com/google/wire"
)

// ProviderSet 数据交换模块依赖集合
var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewDataIOService,
)

func InitModule(attributeModule *attribute.Module, resourceModule *resource.Module, storage *storage.S3Storage,
	modelModule *model.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.FieldsOf(new(*attribute.Module), "Svc"),
		wire.FieldsOf(new(*resource.Module), "Svc"),
		wire.FieldsOf(new(*model.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
