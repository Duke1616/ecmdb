package sorter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockAttribute struct {
	ID      int64
	SortKey int64
}

func (m mockAttribute) GetID() int64      { return m.ID }
func (m mockAttribute) GetSortKey() int64 { return m.SortKey }

func TestReorder_Unchanged(t *testing.T) {
	elems := []mockAttribute{
		{ID: 1, SortKey: 1000},
		{ID: 2, SortKey: 2000},
		{ID: 3, SortKey: 3000},
	}

	// 元素 2 原本在下标 1，拖拽到位置 1，判定为未改变
	plan := Reorder(elems, 2, 1)

	assert.True(t, plan.Unchanged)
	assert.False(t, plan.NeedRebalance)
	assert.Equal(t, int64(2000), plan.NewSortKey)
}

func TestReorder_FastPath(t *testing.T) {
	elems := []mockAttribute{
		{ID: 1, SortKey: 1000},
		{ID: 2, SortKey: 2000},
		{ID: 3, SortKey: 3000},
		{ID: 4, SortKey: 4000},
	}

	// 将元素 2 移动到首位
	plan := Reorder(elems, 2, 0)

	assert.False(t, plan.Unchanged)
	assert.False(t, plan.NeedRebalance)
	assert.Equal(t, int64(500), plan.NewSortKey) // 1000 / 2 = 500
}

func TestReorder_SlowPath_Rebalance(t *testing.T) {
	elems := []mockAttribute{
		{ID: 1, SortKey: 1000},
		{ID: 2, SortKey: 1001},
		{ID: 3, SortKey: 1002},
	}

	// 将元素 3 移动到下标 1（介于 1000 与 1001 之间，间隙不足）
	plan := Reorder(elems, 3, 1)

	assert.False(t, plan.Unchanged)
	assert.True(t, plan.NeedRebalance)
	assert.Len(t, plan.Items, 3)

	assert.Equal(t, Item{ID: 1, SortKey: 1000}, plan.Items[0])
	assert.Equal(t, Item{ID: 3, SortKey: 2000}, plan.Items[1])
	assert.Equal(t, Item{ID: 2, SortKey: 3000}, plan.Items[2])
}

func TestReorder_HeadExhaustionRebalance(t *testing.T) {
	elems := []mockAttribute{
		{ID: 1, SortKey: 1},
		{ID: 2, SortKey: 1000},
	}

	// 插入新元素 99 到首位，1 / 2 = 0 空间耗尽，触发重平衡
	plan := Reorder(elems, 99, 0)

	assert.False(t, plan.Unchanged)
	assert.True(t, plan.NeedRebalance)
	assert.Len(t, plan.Items, 3)
	assert.Equal(t, Item{ID: 99, SortKey: 1000}, plan.Items[0])
	assert.Equal(t, Item{ID: 1, SortKey: 2000}, plan.Items[1])
	assert.Equal(t, Item{ID: 2, SortKey: 3000}, plan.Items[2])
}

func TestReorder_CrossGroup(t *testing.T) {
	targetElems := []mockAttribute{
		{ID: 1, SortKey: 1000},
		{ID: 2, SortKey: 2000},
		{ID: 3, SortKey: 3000},
	}

	// 将外部元素 99 插入到下标 1
	plan := Reorder(targetElems, 99, 1)

	assert.False(t, plan.Unchanged)
	assert.False(t, plan.NeedRebalance)
	assert.Equal(t, int64(1500), plan.NewSortKey)
}

func TestReorder_InsertAtEnd(t *testing.T) {
	elems := []mockAttribute{
		{ID: 1, SortKey: 1000},
		{ID: 2, SortKey: 2000},
	}

	plan := Reorder(elems, 99, 2)

	assert.False(t, plan.Unchanged)
	assert.False(t, plan.NeedRebalance)
	assert.Equal(t, int64(3000), plan.NewSortKey)
}

func TestReorder_EmptyList(t *testing.T) {
	var elems []mockAttribute

	plan := Reorder(elems, 1, 0)

	assert.False(t, plan.Unchanged)
	assert.False(t, plan.NeedRebalance)
	assert.Equal(t, int64(1000), plan.NewSortKey)
}

func TestReorderWithGap(t *testing.T) {
	elems := []mockAttribute{
		{ID: 1, SortKey: 100},
		{ID: 2, SortKey: 101},
		{ID: 3, SortKey: 102},
	}

	// 自定义 gap = 500
	plan := ReorderWithGap(elems, 3, 1, 500)

	assert.True(t, plan.NeedRebalance)
	assert.Equal(t, Item{ID: 1, SortKey: 500}, plan.Items[0])
	assert.Equal(t, Item{ID: 3, SortKey: 1000}, plan.Items[1])
	assert.Equal(t, Item{ID: 2, SortKey: 1500}, plan.Items[2])
}
