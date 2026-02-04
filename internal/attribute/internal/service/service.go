package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/event"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

const (
	// IndexGap 稀疏索引间隔
	IndexGap = 1000
)

//go:generate mockgen -source=./service.go -destination=../../mocks/attribute.mock.go -package=attributemocks -typed Service
type Service interface {
	// CreateAttribute 创建模型字段
	CreateAttribute(ctx context.Context, req domain.Attribute) (int64, error)

	// BatchCreateAttribute 批量创建模型字段
	BatchCreateAttribute(ctx context.Context, attrs []domain.Attribute) error

	// SearchAttributeFieldsByModelUid 查询模型下的所有字段信息，不包含安全字段，内部使用
	SearchAttributeFieldsByModelUid(ctx context.Context, modelUid string) ([]string, error)

	// SearchAllAttributeFieldsByModelUid 查询模型下的所有字段信息，内部使用
	SearchAllAttributeFieldsByModelUid(ctx context.Context, modelUid string) ([]string, error)

	// SearchAttributeFieldsBySecure 查询全有的安全字段
	SearchAttributeFieldsBySecure(ctx context.Context, modelUids []string) (map[string][]string, error)

	// ListAttributes 查询模型下的所有字段详情信息，前端使用
	ListAttributes(ctx context.Context, modelUID string) ([]domain.Attribute, int64, error)

	// DeleteAttribute 删除模型字段
	DeleteAttribute(ctx context.Context, id int64) (int64, error)

	// UpdateAttribute 更新模型字段
	UpdateAttribute(ctx context.Context, attribute domain.Attribute) (int64, error)

	// CustomAttributeFieldColumns 自定义展示字段、以及排序
	CustomAttributeFieldColumns(ctx *gin.Context, modelUid string, customField []string) (int64, error)

	// ListAttributePipeline 根据组聚合获取每个组下的所有字段
	ListAttributePipeline(ctx *gin.Context, modelUid string) ([]domain.AttributePipeline, error)

	// CreateDefaultAttribute 创建新模型，创建默认字段信息
	CreateDefaultAttribute(ctx context.Context, modelUid string) (int64, error)

	// CreateAttributeGroup 创建模型字段组
	CreateAttributeGroup(ctx context.Context, req domain.AttributeGroup) (int64, error)

	// ListAttributeGroup 模型组
	ListAttributeGroup(ctx context.Context, modelUid string) ([]domain.AttributeGroup, error)

	// ListAttributeGroupByIds 返回每个组下面的属性
	ListAttributeGroupByIds(ctx context.Context, ids []int64) ([]domain.AttributeGroup, error)

	// BatchCreateAttributeGroup 批量创建组
	BatchCreateAttributeGroup(ctx context.Context, ags []domain.AttributeGroup) ([]domain.AttributeGroup, error)

	// DeleteAttributeGroup 删除模型字段组
	DeleteAttributeGroup(ctx context.Context, id int64) (int64, error)

	// RenameAttributeGroup 重命名属性组
	RenameAttributeGroup(ctx context.Context, id int64, name string) (int64, error)

	// Sort 属性拖拽排序
	Sort(ctx context.Context, id, targetGroupId, targetPosition int64) error

	// SortAttributeGroup 属性组拖拽排序
	SortAttributeGroup(ctx context.Context, id, targetPosition int64) error
}

type service struct {
	repo      repository.AttributeRepository
	producer  event.FieldSecureAttrChangeEventProducer
	groupRepo repository.AttributeGroupRepository
}

func (s *service) BatchCreateAttributeGroup(ctx context.Context, ags []domain.AttributeGroup) ([]domain.AttributeGroup, error) {
	return s.groupRepo.BatchCreateAttributeGroup(ctx, ags)
}

func (s *service) UpdateAttribute(ctx context.Context, attribute domain.Attribute) (int64, error) {
	// 查看当前的数据
	attr, err := s.repo.DetailAttribute(ctx, attribute.ID)
	if err != nil {
		return 0, err
	}

	// 更新数据
	id, err := s.repo.UpdateAttribute(ctx, attribute)
	if err != nil {
		return 0, err
	}

	// 比对安全属性是否变更
	if attr.Secure == attribute.Secure {
		return id, nil
	}

	// 如果更新了则推送事件
	err = s.producer.Produce(ctx, event.FieldSecureAttrChange{
		ModelUid:   attr.ModelUid,
		FieldUid:   attr.FieldUid,
		Secure:     attribute.Secure,
		TiggerTime: time.Now().UnixMilli(),
	})
	return id, err
}

func NewService(repo repository.AttributeRepository, groupRepo repository.AttributeGroupRepository,
	producer event.FieldSecureAttrChangeEventProducer) Service {
	return &service{
		repo:      repo,
		groupRepo: groupRepo,
		producer:  producer,
	}
}

func (s *service) SearchAllAttributeFieldsByModelUid(ctx context.Context, modelUid string) ([]string, error) {
	attrs, err := s.repo.ListAttributes(ctx, modelUid)
	if err != nil {
		return nil, err
	}

	return slice.Map(attrs, func(idx int, src domain.Attribute) string {
		return src.FieldUid
	}), nil
}

func (s *service) CreateAttribute(ctx context.Context, req domain.Attribute) (int64, error) {
	// NOTE: 分配稀疏索引，防止频繁更新
	if req.SortKey == 0 {
		maxSortKey, err := s.repo.GetMaxSortKeyByGroupID(ctx, req.GroupId)
		if err != nil {
			return 0, err
		}
		req.SortKey = maxSortKey + 1000
	}
	return s.repo.CreateAttribute(ctx, req)
}

func (s *service) BatchCreateAttribute(ctx context.Context, attrs []domain.Attribute) error {
	return s.repo.BatchCreateAttribute(ctx, attrs)
}

func (s *service) SearchAttributeFieldsByModelUid(ctx context.Context, modelUid string) ([]string, error) {
	return s.repo.SearchAttributeFieldsByModelUid(ctx, modelUid)
}

func (s *service) ListAttributes(ctx context.Context, modelUid string) ([]domain.Attribute, int64, error) {
	var (
		total int64
		attrs []domain.Attribute
		eg    errgroup.Group
	)
	eg.Go(func() error {
		var err error
		attrs, err = s.repo.ListAttributes(ctx, modelUid)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx, modelUid)
		return err
	})
	if err := eg.Wait(); err != nil {
		return attrs, total, err
	}
	return attrs, total, nil
}

func (s *service) CustomAttributeFieldColumns(ctx *gin.Context, modelUid string, customField []string) (int64, error) {
	var (
		total int64
		eg    errgroup.Group
	)
	eg.Go(func() error {
		var err error
		total, err = s.repo.CustomAttributeFieldColumns(ctx, modelUid, customField)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.CustomAttributeFieldColumnsReverse(ctx, modelUid, customField)
		return err
	})
	if err := eg.Wait(); err != nil {
		return total, err
	}
	return total, nil
}

func (s *service) DeleteAttribute(ctx context.Context, id int64) (int64, error) {
	attr, err := s.repo.DetailAttribute(ctx, id)
	if err != nil {
		return 0, err
	}
	if attr.Builtin {
		return 0, fmt.Errorf("内置属性不允许删除")
	}
	return s.repo.DeleteAttribute(ctx, id)
}

func (s *service) CreateDefaultAttribute(ctx context.Context, modelUid string) (int64, error) {
	groupId, err := s.CreateAttributeGroup(ctx, domain.AttributeGroup{
		Name:     "基础属性",
		ModelUid: modelUid,
		SortKey:  0,
	})
	if err != nil {
		return 0, err
	}

	attr := s.defaultAttr(modelUid, groupId)

	return s.repo.CreateAttribute(ctx, attr)
}

func (s *service) ListAttributeGroupByIds(ctx context.Context, ids []int64) ([]domain.AttributeGroup, error) {
	return s.groupRepo.ListAttributeGroupByIds(ctx, ids)
}

func (s *service) ListAttributeGroup(ctx context.Context, modelUid string) ([]domain.AttributeGroup, error) {
	return s.groupRepo.ListAttributeGroup(ctx, modelUid)
}

func (s *service) CreateAttributeGroup(ctx context.Context, req domain.AttributeGroup) (int64, error) {
	return s.groupRepo.CreateAttributeGroup(ctx, req)
}

func (s *service) ListAttributePipeline(ctx *gin.Context, modelUid string) ([]domain.AttributePipeline, error) {
	return s.repo.ListAttributePipeline(ctx, modelUid)
}

func (s *service) SearchAttributeFieldsBySecure(ctx context.Context, modelUids []string) (map[string][]string, error) {
	return s.repo.SearchAttributeFieldsBySecure(ctx, modelUids)
}

func (s *service) defaultAttr(modelUid string, groupId int64) domain.Attribute {
	return domain.Attribute{
		ModelUid:  modelUid,
		Index:     0,
		Display:   true,
		Required:  true,
		FieldName: "名称",
		FieldType: "string",
		FieldUid:  "name",
		GroupId:   groupId,
		Secure:    false,
		Builtin:   true,
	}
}

func (s *service) DeleteAttributeGroup(ctx context.Context, id int64) (int64, error) {
	// 1. 删除组下的所有 Attributes
	if _, err := s.repo.DeleteByGroupId(ctx, id); err != nil {
		return 0, err
	}
	// 2. 删除 Group
	return s.groupRepo.DeleteAttributeGroup(ctx, id)
}

func (s *service) RenameAttributeGroup(ctx context.Context, id int64, name string) (int64, error) {
	return s.groupRepo.RenameAttributeGroup(ctx, id, name)
}

// Sort 属性拖拽排序（执行计划模式）
func (s *service) Sort(ctx context.Context, id, targetGroupId, targetPosition int64) error {
	// 1. 获取目标分组的所有属性
	targetAttrs, err := s.repo.ListByGroupID(ctx, targetGroupId)
	if err != nil {
		return err
	}

	// 2. 获取被拖拽的属性（支持跨组）
	var draggedAttr domain.Attribute
	draggedIdx := slices.IndexFunc(targetAttrs, func(a domain.Attribute) bool {
		return a.ID == id
	})

	if draggedIdx == -1 {
		// 跨组：查出详情
		draggedAttr, err = s.repo.DetailAttribute(ctx, id)
		if err != nil {
			return err
		}
		// 修正 GroupId 为目标组
		draggedAttr.GroupId = targetGroupId
	} else {
		draggedAttr = targetAttrs[draggedIdx]
	}

	// 3. 计算重排方案
	plan := s.planReorder(targetAttrs, draggedAttr, targetPosition)

	// 4. 执行计划
	if plan.NeedRebalance {
		// 慢路径：批量更新整个分组
		return s.repo.BatchUpdateSortKey(ctx, plan.Items)
	}

	// 快速路径：单条更新
	return s.repo.UpdateSort(ctx, id, targetGroupId, plan.NewSortKey)
}

// planReorder 计算重排方案（核心算法，纯函数）
func (s *service) planReorder(targetAttrs []domain.Attribute, draggedAttr domain.Attribute, targetPosition int64) domain.ReorderPlan {
	// 1. 移除被拖拽元素（如果是组内拖拽），得到剩余列表
	remainingAttrs := s.removeDraggedAttr(targetAttrs, draggedAttr.ID)

	// 2. 基于剩余列表和目标位置，计算新的 SortKey
	newSortKey := s.calculateSortKey(remainingAttrs, targetPosition)

	// 3. 检测是否需要重平衡
	if s.needsRebalance(remainingAttrs, targetPosition, newSortKey) {
		// 4. 构建包含被拖拽元素的完整最终列表
		finalList := s.insertAttr(remainingAttrs, draggedAttr, targetPosition)

		// 触发重平衡：生成批量更新方案
		return domain.ReorderPlan{
			NeedRebalance: true,
			Items:         s.generateRebalanceItems(finalList),
		}
	}

	// 快速路径：返回单条更新方案
	return domain.ReorderPlan{
		NeedRebalance: false,
		NewSortKey:    newSortKey,
	}
}

// removeDraggedAttr 移除被拖拽元素
func (s *service) removeDraggedAttr(attrs []domain.Attribute, draggedId int64) []domain.Attribute {
	idx := slices.IndexFunc(attrs, func(a domain.Attribute) bool {
		return a.ID == draggedId
	})
	if idx == -1 {
		return attrs
	}
	return slices.Delete(slices.Clone(attrs), idx, idx+1)
}

// insertAttr 将元素插入到指定位置
func (s *service) insertAttr(attrs []domain.Attribute, attr domain.Attribute, position int64) []domain.Attribute {
	// 修正 position 范围
	if position < 0 {
		position = 0
	}
	if position > int64(len(attrs)) {
		position = int64(len(attrs))
	}

	// 插入
	result := slices.Insert(slices.Clone(attrs), int(position), attr)
	// 并不是所有时候都需要 clone，但为了安全起见
	return result
}

// calculateSortKey 计算新的 SortKey（统一算法）
func (s *service) calculateSortKey(attrs []domain.Attribute, position int64) int64 {
	n := int64(len(attrs))

	// 边界：空列表或末尾插入
	if n == 0 || position >= n {
		if n == 0 {
			return IndexGap
		}
		return attrs[n-1].SortKey + IndexGap
	}

	// 开头插入
	if position == 0 {
		return attrs[0].SortKey / 2
	}

	// 中间插入：取前后中点
	return (attrs[position-1].SortKey + attrs[position].SortKey) / 2
}

// needsRebalance 检测是否需要重平衡
func (s *service) needsRebalance(attrs []domain.Attribute, position, newSortKey int64) bool {
	// 只有中间插入才可能冲突
	if position <= 0 || position >= int64(len(attrs)) {
		return false
	}
	// SortKey 冲突（间隙 < 1）
	return newSortKey <= attrs[position-1].SortKey
}

// generateRebalanceItems 生成重平衡的批量更新方案
func (s *service) generateRebalanceItems(attrs []domain.Attribute) []domain.AttributeSortItem {
	return slice.Map(attrs, func(idx int, src domain.Attribute) domain.AttributeSortItem {
		return domain.AttributeSortItem{
			ID:      src.ID,
			GroupId: src.GroupId,
			SortKey: int64(idx+1) * IndexGap,
		}
	})
}

// SortAttributeGroup 属性组拖拽排序（执行计划模式）
func (s *service) SortAttributeGroup(ctx context.Context, id, targetPosition int64) error {
	// 0. 获取当前组信息，拿到 ModelUid
	groups, err := s.groupRepo.ListAttributeGroupByIds(ctx, []int64{id})
	if err != nil {
		return err
	}
	if len(groups) == 0 {
		return fmt.Errorf("属性组不存在")
	}
	modelUid := groups[0].ModelUid

	// 1. 获取该模型下的所有分组（已排序）
	allGroups, err := s.groupRepo.ListAttributeGroup(ctx, modelUid)
	if err != nil {
		return err
	}

	// 2. 获取被拖拽的分组
	draggedIdx := slices.IndexFunc(allGroups, func(g domain.AttributeGroup) bool {
		return g.ID == id
	})
	if draggedIdx == -1 {
		return fmt.Errorf("被拖拽的分组未在列表中找到")
	}
	draggedGroup := allGroups[draggedIdx]

	// 3. 计算重排方案
	plan := s.planReorderGroup(allGroups, draggedGroup, targetPosition)

	// 4. 执行计划
	if plan.NeedRebalance {
		// 慢路径：批量更新整个模型下的分组
		return s.groupRepo.BatchUpdateSort(ctx, plan.Items)
	}

	// 快速路径：单条更新
	return s.groupRepo.UpdateSort(ctx, id, plan.NewSortKey)
}

// planReorderGroup 计算属性组重排方案（核心算法，纯函数）
func (s *service) planReorderGroup(groups []domain.AttributeGroup, draggedGroup domain.AttributeGroup, targetPosition int64) domain.ReorderGroupPlan {
	// 1. 移除被拖拽元素
	remainingGroups := s.removeDraggedGroup(groups, draggedGroup.ID)

	// 2. 计算新的 SortKey
	newSortKey := s.calculateSortKeyGroup(remainingGroups, targetPosition)

	// 3. 检测是否需要重平衡
	if s.needsRebalanceGroup(remainingGroups, targetPosition, newSortKey) {
		// 4. 构建完整列表
		finalList := s.insertGroup(remainingGroups, draggedGroup, targetPosition)

		// 触发重平衡：生成批量更新方案
		return domain.ReorderGroupPlan{
			NeedRebalance: true,
			Items:         s.generateRebalanceItemsGroup(finalList),
		}
	}

	// 快速路径：返回单条更新方案
	return domain.ReorderGroupPlan{
		NeedRebalance: false,
		NewSortKey:    newSortKey,
	}
}

// removeDraggedGroup 移除被拖拽分组
func (s *service) removeDraggedGroup(groups []domain.AttributeGroup, draggedId int64) []domain.AttributeGroup {
	idx := slices.IndexFunc(groups, func(g domain.AttributeGroup) bool {
		return g.ID == draggedId
	})
	if idx == -1 {
		return groups
	}
	return slices.Delete(slices.Clone(groups), idx, idx+1)
}

// insertGroup 插入分组
func (s *service) insertGroup(groups []domain.AttributeGroup, group domain.AttributeGroup, position int64) []domain.AttributeGroup {
	if position < 0 {
		position = 0
	}
	if position > int64(len(groups)) {
		position = int64(len(groups))
	}
	return slices.Insert(slices.Clone(groups), int(position), group)
}

// calculateSortKeyGroup 计算新的 SortKey (属性组版本)
func (s *service) calculateSortKeyGroup(groups []domain.AttributeGroup, position int64) int64 {
	n := int64(len(groups))

	// 边界：空列表或末尾插入
	if n == 0 || position >= n {
		if n == 0 {
			return IndexGap
		}
		return groups[n-1].SortKey + IndexGap
	}

	// 开头插入
	if position == 0 {
		return groups[0].SortKey / 2
	}

	// 中间插入：取前后中点
	return (groups[position-1].SortKey + groups[position].SortKey) / 2
}

// needsRebalanceGroup 检测是否需要重平衡 (属性组版本)
func (s *service) needsRebalanceGroup(groups []domain.AttributeGroup, position, newSortKey int64) bool {
	// 只有中间插入才可能冲突
	if position <= 0 || position >= int64(len(groups)) {
		return false
	}
	// SortKey 冲突（间隙 < 1）
	return newSortKey <= groups[position-1].SortKey
}

// generateRebalanceItemsGroup 生成重平衡的批量更新方案 (属性组版本)
func (s *service) generateRebalanceItemsGroup(groups []domain.AttributeGroup) []domain.AttributeGroupSortItem {
	return slice.Map(groups, func(idx int, src domain.AttributeGroup) domain.AttributeGroupSortItem {
		return domain.AttributeGroupSortItem{
			ID:      src.ID,
			SortKey: int64(idx+1) * IndexGap,
		}
	})
}
