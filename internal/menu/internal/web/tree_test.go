package web

import (
	"testing"

	"github.com/Duke1616/ecmdb/internal/menu/internal/domain"
)

func TestGetMenusTree(t *testing.T) {
	tests := []struct {
		name     string
		menus    []domain.Menu
		expected int // 期望的根节点数量
	}{
		{
			name: "正常树结构",
			menus: []domain.Menu{
				{Id: 1, Pid: 0, Name: "根节点1"},
				{Id: 2, Pid: 1, Name: "子节点1"},
				{Id: 3, Pid: 1, Name: "子节点2"},
			},
			expected: 1,
		},
		{
			name: "孤儿节点作为根节点",
			menus: []domain.Menu{
				{Id: 1, Pid: 0, Name: "根节点1"},
				{Id: 2, Pid: 1, Name: "子节点1"},
				{Id: 3, Pid: 999, Name: "孤儿节点"}, // 父节点999不存在
			},
			expected: 2, // 根节点1 + 孤儿节点
		},
		{
			name: "延迟关联的节点",
			menus: []domain.Menu{
				{Id: 1, Pid: 0, Name: "根节点1"},
				{Id: 2, Pid: 3, Name: "子节点2"}, // 父节点3在后面定义
				{Id: 3, Pid: 1, Name: "子节点1"},
			},
			expected: 1, // 只有根节点1，子节点2应该被正确关联到子节点1下
		},
		{
			name: "多个孤儿节点",
			menus: []domain.Menu{
				{Id: 1, Pid: 0, Name: "根节点1"},
				{Id: 2, Pid: 999, Name: "孤儿节点1"}, // 父节点999不存在
				{Id: 3, Pid: 888, Name: "孤儿节点2"}, // 父节点888不存在
			},
			expected: 3, // 根节点1 + 两个孤儿节点
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMenusTree(tt.menus)
			if len(result) != tt.expected {
				t.Errorf("GetMenusTree() 根节点数量 = %v, 期望 %v", len(result), tt.expected)
			}
			
			// 打印结果用于调试
			t.Logf("测试 %s: 根节点数量 = %d", tt.name, len(result))
			for i, root := range result {
				t.Logf("  根节点 %d: ID=%d, Name=%s, 子节点数量=%d", i+1, root.Id, root.Name, len(root.Children))
			}
		})
	}
}