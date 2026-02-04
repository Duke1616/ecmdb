package sorter

import (
	"slices"

	"github.com/ecodeclub/ekit/slice"
)

const (
	// DefaultIndexGap 默认稀疏索引间隔
	DefaultIndexGap = 1000
)

// Sortable 可排序元素接口
// NOTE: 所有需要使用拖拽排序功能的实体都应实现此接口
type Sortable interface {
	GetID() int64
	GetSortKey() int64
}

// SortItem 排序更新项接口
// NOTE: 用于批量更新时的数据传输
type SortItem interface {
	any
}

// ReorderPlan 重排执行计划
// NOTE: T 为排序项类型,必须满足 SortItem 约束
type ReorderPlan[T SortItem] struct {
	// NeedRebalance 是否需要重平衡
	NeedRebalance bool
	// NewSortKey 单个元素的新 SortKey（快速路径）
	NewSortKey int64
	// Items 批量更新的元素列表（慢路径）
	Items []T
}

// Sorter 通用排序器
// NOTE: E 为元素类型(实现 Sortable), T 为排序项类型(实现 SortItem)
type Sorter[E Sortable, T SortItem] struct {
	indexGap int64
	// convertFunc 将元素转换为排序项的函数
	convertFunc func(elem E, idx int) T
}

// NewSorter 创建排序器
// convertFunc: 将元素转换为排序项,用于重平衡场景
// 例如: func(attr Attribute, idx int) AttributeSortItem { ... }
func NewSorter[E Sortable, T SortItem](convertFunc func(elem E, idx int) T) *Sorter[E, T] {
	return &Sorter[E, T]{
		indexGap:    DefaultIndexGap,
		convertFunc: convertFunc,
	}
}

// WithIndexGap 设置索引间隔
func (s *Sorter[E, T]) WithIndexGap(gap int64) *Sorter[E, T] {
	s.indexGap = gap
	return s
}

// PlanReorder 计算重排方案（核心算法，纯函数）
// elements: 目标分组/列表中的所有元素
// draggedElem: 被拖拽的元素
// targetPosition: 目标位置 (0-based)
func (s *Sorter[E, T]) PlanReorder(elements []E, draggedElem E, targetPosition int64) ReorderPlan[T] {
	// 1. 移除被拖拽元素（如果是组内拖拽），得到剩余列表
	remainingElems := s.removeDragged(elements, draggedElem.GetID())

	// 2. 基于剩余列表和目标位置，计算新的 SortKey
	newSortKey := s.calculateSortKey(remainingElems, targetPosition)

	// 3. 检测是否需要重平衡
	if s.needsRebalance(remainingElems, targetPosition, newSortKey) {
		// 4. 构建包含被拖拽元素的完整最终列表
		finalList := s.insertElem(remainingElems, draggedElem, targetPosition)

		// 触发重平衡：生成批量更新方案
		return ReorderPlan[T]{
			NeedRebalance: true,
			Items:         s.generateRebalanceItems(finalList),
		}
	}

	// 快速路径：返回单条更新方案
	return ReorderPlan[T]{
		NeedRebalance: false,
		NewSortKey:    newSortKey,
	}
}

// removeDragged 移除被拖拽元素
func (s *Sorter[E, T]) removeDragged(elems []E, draggedId int64) []E {
	idx := slices.IndexFunc(elems, func(e E) bool {
		return e.GetID() == draggedId
	})
	if idx == -1 {
		return elems
	}
	return slices.Delete(slices.Clone(elems), idx, idx+1)
}

// insertElem 将元素插入到指定位置
func (s *Sorter[E, T]) insertElem(elems []E, elem E, position int64) []E {
	// 修正 position 范围
	if position < 0 {
		position = 0
	}
	if position > int64(len(elems)) {
		position = int64(len(elems))
	}

	// 插入
	result := slices.Insert(slices.Clone(elems), int(position), elem)
	return result
}

// calculateSortKey 计算新的 SortKey（统一算法）
func (s *Sorter[E, T]) calculateSortKey(elems []E, position int64) int64 {
	n := int64(len(elems))

	// 边界：空列表或末尾插入
	if n == 0 || position >= n {
		if n == 0 {
			return s.indexGap
		}
		return elems[n-1].GetSortKey() + s.indexGap
	}

	// 开头插入
	if position == 0 {
		return elems[0].GetSortKey() / 2
	}

	// 中间插入：取前后中点
	return (elems[position-1].GetSortKey() + elems[position].GetSortKey()) / 2
}

// needsRebalance 检测是否需要重平衡
func (s *Sorter[E, T]) needsRebalance(elems []E, position, newSortKey int64) bool {
	// 只有中间插入才可能冲突
	if position <= 0 || position >= int64(len(elems)) {
		return false
	}
	// SortKey 冲突（间隙 < 1）
	return newSortKey <= elems[position-1].GetSortKey()
}

// generateRebalanceItems 生成重平衡的批量更新方案
func (s *Sorter[E, T]) generateRebalanceItems(elems []E) []T {
	return slice.Map(elems, func(idx int, src E) T {
		return s.convertFunc(src, idx)
	})
}
