package executor

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

// AttributeExecutor 属性执行器，负责属性分组和字段的创建
type AttributeExecutor interface {
	// ExecuteGroups 执行属性分组创建，直接使用 domain.AttributeGroup
	ExecuteGroups(ctx context.Context, modelUid string, groups []attribute.AttributeGroup) (
		[]attribute.AttributeGroup, error)
	
	// ExecuteFields 执行字段创建，直接使用 domain.Attribute
	ExecuteFields(ctx context.Context, modelUid string, attrs []attribute.Attribute) error
}

type attributeExecutor struct {
	attributeSvc attribute.Service
	logger       *elog.Component
}

// NewAttributeExecutor 创建属性执行器
func NewAttributeExecutor(attributeSvc attribute.Service) AttributeExecutor {
	return &attributeExecutor{
		attributeSvc: attributeSvc,
		logger:       elog.DefaultLogger,
	}
}

// ExecuteGroups 执行属性分组创建
func (e *attributeExecutor) ExecuteGroups(ctx context.Context, modelUid string, groups []attribute.AttributeGroup) ([]attribute.AttributeGroup, error) {
	if len(groups) == 0 {
		e.logger.Info("无属性分组需要创建")
		return nil, nil
	}

	// 获取模型属性分组
	existingGroups, err := e.attributeSvc.ListAttributeGroup(ctx, modelUid)
	if err != nil {
		return nil, err
	}

	// 创建已存在分组的名称映射，用于快速查找
	existingMap := make(map[string]struct{})
	for _, group := range existingGroups {
		existingMap[group.Name] = struct{}{}
	}

	// 使用 FilterMap 找出需要创建的分组（数据库中不存在的）
	groupsToCreate := slice.FilterMap(groups, func(idx int, src attribute.AttributeGroup) (attribute.AttributeGroup, bool) {
		// 如果分组不存在于数据库中，返回 true
		_, exists := existingMap[src.Name]
		return src, !exists
	})

	// 如果没有需要创建的分组，直接返回已存在的分组
	if len(groupsToCreate) == 0 {
		e.logger.Info("所有模型属性分组已存在，无需创建")
		return existingGroups, nil
	}

	// 创建不存的分组
	newGroups, err := e.attributeSvc.BatchCreateAttributeGroup(ctx, groupsToCreate)
	if err != nil {
		return nil, err
	}

	e.logger.Info("模型属性分组创建完成", elog.Int("分组数量", len(newGroups)))
	return append(existingGroups, newGroups...), nil
}

// ExecuteFields 执行字段创建
func (e *attributeExecutor) ExecuteFields(ctx context.Context, modelUid string, attrs []attribute.Attribute) error {
	if len(attrs) == 0 {
		return nil
	}

	// 获取模型属性
	existingAttrs, _, err := e.attributeSvc.ListAttributes(ctx, modelUid)
	if err != nil {
		return err
	}

	// 创建已存在分组的名称映射，用于快速查找
	existingMap := make(map[string]struct{})
	for _, attr := range existingAttrs {
		existingMap[attr.FieldUid] = struct{}{}
	}

	// 使用 FilterMap 找出需要创建的分组（数据库中不存在的）
	attrsToCreate := slice.FilterMap(attrs, func(idx int, src attribute.Attribute) (attribute.Attribute, bool) {
		// 如果分组不存在于数据库中，返回 true
		_, exists := existingMap[src.FieldUid]
		return src, !exists
	})

	// 如果没有需要创建的分组，直接返回已存在的分组
	if len(attrsToCreate) == 0 {
		e.logger.Info("所有模型属性已存在，无需创建")
		return nil
	}

	// 创建不存的分组
	return e.attributeSvc.BatchCreateAttribute(ctx, attrsToCreate)
}
