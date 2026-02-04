package sorter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockAttribute 模拟属性实体
type mockAttribute struct {
	ID      int64
	GroupId int64
	SortKey int64
}

func (m mockAttribute) GetID() int64      { return m.ID }
func (m mockAttribute) GetSortKey() int64 { return m.SortKey }

// mockAttributeSortItem 模拟排序项
type mockAttributeSortItem struct {
	ID      int64
	GroupId int64
	SortKey int64
}

// TestSorter_PlanReorder_FastPath 测试快速路径（无需重平衡）
func TestSorter_PlanReorder_FastPath(t *testing.T) {
	// 准备测试数据
	elems := []mockAttribute{
		{ID: 1, SortKey: 1000},
		{ID: 2, SortKey: 2000},
		{ID: 3, SortKey: 3000},
		{ID: 4, SortKey: 4000},
	}

	draggedElem := mockAttribute{ID: 2, SortKey: 2000}

	// 创建排序器
	sorter := NewSorter[mockAttribute, mockAttributeSortItem](
		func(elem mockAttribute, idx int) mockAttributeSortItem {
			return mockAttributeSortItem{
				ID:      elem.ID,
				GroupId: elem.GroupId,
				SortKey: int64(idx+1) * DefaultIndexGap,
			}
		},
	)

	// 测试：将第2个元素移动到第1个位置
	plan := sorter.PlanReorder(elems, draggedElem, 0)

	// 验证：应该是快速路径
	assert.False(t, plan.NeedRebalance)
	assert.Equal(t, int64(500), plan.NewSortKey) // 1000 / 2 = 500
}

// TestSorter_PlanReorder_SlowPath 测试慢路径（需要重平衡）
func TestSorter_PlanReorder_SlowPath(t *testing.T) {
	// 准备测试数据: SortKey 间隙很小
	elems := []mockAttribute{
		{ID: 1, SortKey: 1000},
		{ID: 2, SortKey: 1001}, // 间隙只有1
		{ID: 3, SortKey: 1002},
	}

	draggedElem := mockAttribute{ID: 3, SortKey: 1002}

	// 创建排序器
	sorter := NewSorter[mockAttribute, mockAttributeSortItem](
		func(elem mockAttribute, idx int) mockAttributeSortItem {
			return mockAttributeSortItem{
				ID:      elem.ID,
				GroupId: elem.GroupId,
				SortKey: int64(idx+1) * DefaultIndexGap,
			}
		},
	)

	// 测试：将第3个元素移动到第2个位置
	plan := sorter.PlanReorder(elems, draggedElem, 1)

	// 验证：应该触发重平衡
	assert.True(t, plan.NeedRebalance)
	assert.Len(t, plan.Items, 3)

	// 验证重平衡后的 SortKey
	assert.Equal(t, int64(1000), plan.Items[0].SortKey)
	assert.Equal(t, int64(2000), plan.Items[1].SortKey)
	assert.Equal(t, int64(3000), plan.Items[2].SortKey)
}

// TestSorter_PlanReorder_CrossGroup 测试跨组拖拽
func TestSorter_PlanReorder_CrossGroup(t *testing.T) {
	// 目标分组的元素
	targetElems := []mockAttribute{
		{ID: 1, GroupId: 100, SortKey: 1000},
		{ID: 2, GroupId: 100, SortKey: 2000},
		{ID: 3, GroupId: 100, SortKey: 3000},
	}

	// 被拖拽的元素（来自另一个分组）
	draggedElem := mockAttribute{ID: 99, GroupId: 200, SortKey: 5000}

	// 创建排序器
	sorter := NewSorter[mockAttribute, mockAttributeSortItem](
		func(elem mockAttribute, idx int) mockAttributeSortItem {
			return mockAttributeSortItem{
				ID:      elem.ID,
				GroupId: elem.GroupId,
				SortKey: int64(idx+1) * DefaultIndexGap,
			}
		},
	)

	// 测试：将外部元素插入到位置1
	plan := sorter.PlanReorder(targetElems, draggedElem, 1)

	// 验证：应该是快速路径
	assert.False(t, plan.NeedRebalance)
	assert.Equal(t, int64(1500), plan.NewSortKey) // (1000 + 2000) / 2 = 1500
}

// TestSorter_PlanReorder_InsertAtEnd 测试插入到末尾
func TestSorter_PlanReorder_InsertAtEnd(t *testing.T) {
	elems := []mockAttribute{
		{ID: 1, SortKey: 1000},
		{ID: 2, SortKey: 2000},
	}

	draggedElem := mockAttribute{ID: 99, SortKey: 5000}

	sorter := NewSorter[mockAttribute, mockAttributeSortItem](
		func(elem mockAttribute, idx int) mockAttributeSortItem {
			return mockAttributeSortItem{
				ID:      elem.ID,
				GroupId: elem.GroupId,
				SortKey: int64(idx+1) * DefaultIndexGap,
			}
		},
	)

	// 测试：插入到末尾
	plan := sorter.PlanReorder(elems, draggedElem, 2)

	// 验证
	assert.False(t, plan.NeedRebalance)
	assert.Equal(t, int64(3000), plan.NewSortKey) // 2000 + 1000 = 3000
}

// TestSorter_PlanReorder_EmptyList 测试空列表
func TestSorter_PlanReorder_EmptyList(t *testing.T) {
	var elems []mockAttribute

	draggedElem := mockAttribute{ID: 1, SortKey: 0}

	sorter := NewSorter[mockAttribute, mockAttributeSortItem](
		func(elem mockAttribute, idx int) mockAttributeSortItem {
			return mockAttributeSortItem{
				ID:      elem.ID,
				GroupId: elem.GroupId,
				SortKey: int64(idx+1) * DefaultIndexGap,
			}
		},
	)

	// 测试：插入到空列表
	plan := sorter.PlanReorder(elems, draggedElem, 0)

	// 验证
	assert.False(t, plan.NeedRebalance)
	assert.Equal(t, int64(1000), plan.NewSortKey) // DefaultIndexGap
}

// TestSorter_WithIndexGap 测试自定义索引间隔
func TestSorter_WithIndexGap(t *testing.T) {
	var elems []mockAttribute
	draggedElem := mockAttribute{ID: 1, SortKey: 0}

	sorter := NewSorter[mockAttribute, mockAttributeSortItem](
		func(elem mockAttribute, idx int) mockAttributeSortItem {
			return mockAttributeSortItem{
				ID:      elem.ID,
				GroupId: elem.GroupId,
				SortKey: int64(idx+1) * 500, // 注意这里的 gap 需要和 WithIndexGap 一致
			}
		},
	).WithIndexGap(500)

	plan := sorter.PlanReorder(elems, draggedElem, 0)

	assert.False(t, plan.NeedRebalance)
	assert.Equal(t, int64(500), plan.NewSortKey)
}
