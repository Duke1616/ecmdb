package incr

import (
	"sort"
	"testing"

	"github.com/Duke1616/ecmdb/cmd/initial/incr/version"
)

func TestVersionComparison(t *testing.T) {
	testCases := []struct {
		name     string
		v1       string
		v2       string
		expected bool
	}{
		{"v1.2.3 > v1.2.2", "v1.2.3", "v1.2.2", true},
		{"v1.2.3 > v1.1.9", "v1.2.3", "v1.1.9", true},
		{"v1.9.2 > v1.2.3", "v1.9.2", "v1.2.3", true},
		{"v1.2.3 = v1.2.3", "v1.2.3", "v1.2.3", false}, // version.Compare 函数返回 false 当相等时
		{"v1.2.2 < v1.2.3", "v1.2.2", "v1.2.3", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v1Slice := version.ParseVersion(tc.v1)
			v2Slice := version.ParseVersion(tc.v2)
			result := version.Compare(v1Slice, v2Slice)
			if result != tc.expected {
				t.Errorf("version.Compare(%s, %s) = %v, expected %v", tc.v1, tc.v2, result, tc.expected)
			}
		})
	}
}

func TestVersionEqual(t *testing.T) {
	testCases := []struct {
		name     string
		v1       string
		v2       string
		expected bool
	}{
		{"v1.2.3 = v1.2.3", "v1.2.3", "v1.2.3", true},
		{"v1.2.3 != v1.2.2", "v1.2.3", "v1.2.2", false},
		{"v1.9.2 = v1.9.2", "v1.9.2", "v1.9.2", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v1Slice := version.ParseVersion(tc.v1)
			v2Slice := version.ParseVersion(tc.v2)
			result := version.Equal(v1Slice, v2Slice)
			if result != tc.expected {
				t.Errorf("version.Equal(%s, %s) = %v, expected %v", tc.v1, tc.v2, result, tc.expected)
			}
		})
	}
}

func TestVersionSorting(t *testing.T) {
	// 模拟注册表中的版本
	versions := []string{"v1.9.2", "v1.2.3", "v1.5.0", "v1.3.0"}
	
	// 手动排序测试
	sort.SliceStable(versions, func(i, j int) bool {
		v1 := version.ParseVersion(versions[i])
		v2 := version.ParseVersion(versions[j])
		return version.Compare(v2, v1) // 升序排序
	})
	
	expected := []string{"v1.2.3", "v1.3.0", "v1.5.0", "v1.9.2"}
	
	for i, version := range versions {
		if version != expected[i] {
			t.Errorf("排序错误: 位置 %d, 得到 %s, 期望 %s", i, version, expected[i])
		}
	}
}
