package executor

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

// RelationExecutor 关联执行器，负责关联类型和模型关联的创建
type RelationExecutor interface {
	// ExecuteRelationTypes 执行关联类型创建，直接使用 domain.RelationType
	ExecuteRelationTypes(ctx context.Context, relationTypes []relation.RelationType) error

	// ExecuteModelRelations 执行模型关联关系创建，直接使用 domain.ModelRelation
	ExecuteModelRelations(ctx context.Context, modelRelations []relation.ModelRelation) error
}

type relationExecutor struct {
	relationRTSvc relation.RTSvc
	relationRMSvc relation.RMSvc
	logger        *elog.Component
}

// NewRelationExecutor 创建关联执行器
func NewRelationExecutor(
	relationRTSvc relation.RTSvc,
	relationRMSvc relation.RMSvc,
) RelationExecutor {
	return &relationExecutor{
		relationRTSvc: relationRTSvc,
		relationRMSvc: relationRMSvc,
		logger:        elog.DefaultLogger,
	}
}

// ExecuteRelationTypes 执行关联类型创建
func (e *relationExecutor) ExecuteRelationTypes(ctx context.Context, relationTypes []relation.RelationType) error {
	if len(relationTypes) == 0 {
		e.logger.Info("无关联类型需要创建")
		return nil
	}

	uids := slice.Map(relationTypes, func(idx int, src relation.RelationType) string {
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
	rtsToCreate := slice.FilterMap(relationTypes, func(idx int, src relation.RelationType) (relation.RelationType, bool) {
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
func (e *relationExecutor) ExecuteModelRelations(ctx context.Context, modelRelations []relation.ModelRelation) error {
	if len(modelRelations) == 0 {
		e.logger.Info("无模型关联关系需要创建")
		return nil
	}

	rms := slice.Map(modelRelations, func(idx int, src relation.ModelRelation) string {
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
	rmsToCreate := slice.FilterMap(modelRelations, func(idx int, src relation.ModelRelation) (relation.ModelRelation, bool) {
		// 如果分组不存在于数据库中，返回 true
		_, exists := existingMap[src.RelationName]
		return src, !exists
	})

	// 如果没有需要创建的分组，直接返回已存在的分组
	if len(rmsToCreate) == 0 {
		e.logger.Info("所有关联类型已存在，无需创建")
		return nil
	}

	e.logger.Info("模型关联关系创建成功")
	return nil
}
