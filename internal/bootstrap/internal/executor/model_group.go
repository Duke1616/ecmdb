package executor

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

// ModelGroupExecutor 模型分组执行器，负责模型分组的创建
type ModelGroupExecutor interface {
	// Execute 执行模型分组创建，直接使用 domain.ModelGroup
	Execute(ctx context.Context, groups []model.ModelGroup) ([]model.ModelGroup, error)

	// Get 查询模型组信息
	Get(ctx context.Context, groupName string) (model.ModelGroup, error)
}

type modelGroupExecutor struct {
	modelGroupSvc model.MGService
	logger        *elog.Component
}

// NewModelGroupExecutor 创建模型分组执行器
func NewModelGroupExecutor(modelGroupSvc model.MGService) ModelGroupExecutor {
	return &modelGroupExecutor{
		modelGroupSvc: modelGroupSvc,
		logger:        elog.DefaultLogger,
	}
}

func (e *modelGroupExecutor) Get(ctx context.Context, groupName string) (model.ModelGroup, error) {
	return e.modelGroupSvc.GetByName(ctx, groupName)
}

// Execute 执行模型分组创建
func (e *modelGroupExecutor) Execute(ctx context.Context, groups []model.ModelGroup) ([]model.ModelGroup, error) {
	if len(groups) == 0 {
		e.logger.Info("无模型分组需要创建")
		return nil, nil
	}

	e.logger.Info("开始创建模型分组", elog.Int("分组数量", len(groups)))

	// 获取需要新增的组名称
	names := slice.Map(groups, func(idx int, src model.ModelGroup) string {
		return src.Name
	})

	// 查询数据库中是否存在
	existingGroups, err := e.modelGroupSvc.GetByNames(ctx, names)
	if err != nil {
		return nil, err
	}

	// 创建已存在分组的名称映射，用于快速查找
	existingMap := make(map[string]struct{})
	for _, group := range existingGroups {
		existingMap[group.Name] = struct{}{}
	}

	// 使用 FilterMap 找出需要创建的分组（数据库中不存在的）
	groupsToCreate := slice.FilterMap(groups, func(idx int, src model.ModelGroup) (model.ModelGroup, bool) {
		// 如果分组不存在于数据库中，返回 true
		_, exists := existingMap[src.Name]
		return src, !exists
	})

	// 如果没有需要创建的分组，直接返回已存在的分组
	if len(groupsToCreate) == 0 {
		e.logger.Info("所有模型分组已存在，无需创建")
		return existingGroups, nil
	}

	// 创建不存的分组
	newGroups, err := e.modelGroupSvc.BatchCreate(ctx, groupsToCreate)
	if err != nil {
		return nil, err
	}

	e.logger.Info("模型分组创建完成", elog.Int("分组数量", len(newGroups)))
	return append(existingGroups, newGroups...), nil
}
