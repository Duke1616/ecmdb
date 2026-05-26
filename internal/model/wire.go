//go:build wireinject

package model

import (
	"context"
	"sync"

	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/model/internal/service"
	"github.com/Duke1616/ecmdb/internal/model/internal/web"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	initMGProvider,
	initModelProvider)

func InitModule(db *mongox.DB, rmModule *relation.Module, attrModule *attribute.Module, resourceSvc *resource.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitModelDAO,
		newResourceDeleteChecker,
		newRelationDeleteChecker,
		initCheckers,
		newAttrCreatorAdapter,
		wire.FieldsOf(new(*relation.Module), "RMSvc"),
		wire.FieldsOf(new(*attribute.Module), "Svc"),
		wire.FieldsOf(new(*resource.Module), "EncryptedSvc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var daoOnce = sync.Once{}

func InitCollectionOnce(db *mongox.DB) {
	daoOnce.Do(func() {
		err := dao.InitIndexes(db)
		if err != nil {
			panic(err)
		}
	})
}

func InitModelDAO(db *mongox.DB) dao.ModelDAO {
	InitCollectionOnce(db)
	return dao.NewModelDAO(db)
}

var initMGProvider = wire.NewSet(
	service.NewMGService,
	repository.NewMGRepository,
	dao.NewModelGroupDAO,
)

var initModelProvider = wire.NewSet(
	service.NewModelService,
	repository.NewModelRepository,
)

type resourceDeleteChecker struct {
	svc resource.Service
}

func (c *resourceDeleteChecker) CheckBeforeDelete(ctx context.Context, modelUid string) error {
	return c.svc.CheckBeforeDelete(ctx, modelUid)
}

func newResourceDeleteChecker(m *resource.Module) *resourceDeleteChecker {
	return &resourceDeleteChecker{svc: m.Svc}
}

type relationDeleteChecker struct {
	svc relation.RMSvc
}

func (c *relationDeleteChecker) CheckBeforeDelete(ctx context.Context, modelUid string) error {
	return c.svc.CheckBeforeDelete(ctx, modelUid)
}

func newRelationDeleteChecker(m *relation.Module) *relationDeleteChecker {
	return &relationDeleteChecker{svc: m.RMSvc}
}

func initCheckers(res *resourceDeleteChecker, rel *relationDeleteChecker) []service.IDeleteModelDependencyChecker {
	return []service.IDeleteModelDependencyChecker{res, rel}
}

// attrCreatorAdapter 将 attribute.Service 适配为 service.IDefaultAttributeCreator
// NOTE: 接口反转适配器——避免 model 模块直接依赖 attribute 模块的具体类型
type attrCreatorAdapter struct {
	svc attribute.Service
}

func (a *attrCreatorAdapter) CreateDefaultAttribute(ctx context.Context, modelUid string) (int64, error) {
	return a.svc.CreateDefaultAttribute(ctx, modelUid)
}

func newAttrCreatorAdapter(svc attribute.Service) service.IDefaultAttributeCreator {
	return &attrCreatorAdapter{svc: svc}
}
