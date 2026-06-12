package executor

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/domain"
	relation "github.com/Duke1616/ecmdb/internal/service/relation"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

// RelationExecutor 关联执行器，负责关联类型和模型关联的创建
type RelationExecutor interface {
	// ExecuteRelationTypes 执行关联类型创建，直接使用 domain.RelationType
	ExecuteRelationTypes(ctx context.Context, relationTypes []domain.RelationType) error

	// ExecuteModelRelations 执行模型关联关系创建，直接使用 domain.ModelRelation
	ExecuteModelRelations(ctx context.Context, modelRelations []domain.ModelRelation) error
}

type relationExecutor struct {
	relationRTSvc relation.RelationTypeService
	relationRMSvc relation.RelationModelService
	logger        *elog.Component
}

// NewRelationExecutor 创建关联执行器
func NewRelationExecutor(
	relationRTSvc relation.RelationTypeService,
	relationRMSvc relation.RelationModelService,
) RelationExecutor {
	return &relationExecutor{
		relationRTSvc: relationRTSvc,
		relationRMSvc: relationRMSvc,
		logger:        elog.DefaultLogger,
	}
}

// ExecuteRelationTypes 执行关联类型创建
func (e *relationExecutor) ExecuteRelationTypes(ctx context.Context, relationTypes []domain.RelationType) error {
	if len(relationTypes) == 0 {
		e.logger.Info("无关联类型需要创建")
		return nil
	}

	uids := slice.Map(relationTypes, func(idx int, src domain.RelationType) string {
		return src.UID
	})

	existingRts, err := e.relationRTSvc.GetByUids(ctx, uids)
	if err != nil {
		return err
	}

	// 创建已存在分组的名称映射，用于快速查找
	existingMap := make(map[string]struct{})
	for _, group := range existingRts {
		existingMap[group.UID] = struct{}{}
	}

	// 使用 FilterMap 找出需要创建的分组（数据库中不存在的）
	rtsToCreate := slice.FilterMap(relationTypes, func(idx int, src domain.RelationType) (domain.RelationType, bool) {
		// 如果分组不存在于数据库中，返回 true
		_, exists := existingMap[src.UID]
		return src, !exists
	})

	// 如果没有需要创建的分组，直接返回已存在的分组
	if len(rtsToCreate) == 0 {
		e.logger.Info("所有关联类型已存在，无需创建")
		return nil
	}

	return e.relationRTSvc.BatchCreate(ctx, rtsToCreate)
}

// ExecuteModelRelations 执行模型关联关系创建
func (e *relationExecutor) ExecuteModelRelations(ctx context.Context, modelRelations []domain.ModelRelation) error {
	if len(modelRelations) == 0 {
		e.logger.Info("无模型关联关系需要创建")
		return nil
	}

	rms := slice.Map(modelRelations, func(idx int, src domain.ModelRelation) string {
		return src.RM()
	})

	existingRMs, err := e.relationRMSvc.GetByRelationNames(ctx, rms)
	if err != nil {
		return err
	}

	// 创建已存在分组的名称映射，用于快速查找
	existingMap := make(map[string]struct{})
	for i := range existingRMs {
		existingMap[existingRMs[i].RelationName] = struct{}{}
	}

	// 使用 FilterMap 找出需要创建的分组（数据库中不存在的）
	rmsToCreate := slice.FilterMap(modelRelations, func(idx int, src domain.ModelRelation) (domain.ModelRelation, bool) {
		// 如果分组不存在于数据库中，返回 true
		_, exists := existingMap[src.RelationName]
		return src, !exists
	})

	// 如果没有需要创建的分组，直接返回已存在的分组
	if len(rmsToCreate) == 0 {
		e.logger.Info("所有关联类型已存在，无需创建")
		return nil
	}

	return e.relationRMSvc.BatchCreate(ctx, rmsToCreate)
}
