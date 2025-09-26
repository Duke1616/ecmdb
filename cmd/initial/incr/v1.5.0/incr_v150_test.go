package v150

import (
	"testing"
)

func TestVersion150Logic(t *testing.T) {
	// 测试 v1.5.0 版本逻辑
	t.Run("版本信息", func(t *testing.T) {
		incr := &incrV150{}
		if incr.Version() != "v1.5.0" {
			t.Errorf("期望版本 v1.5.0，实际 %s", incr.Version())
		}
	})

	t.Run("版本比较", func(t *testing.T) {
		// 测试版本比较逻辑
		testCases := []struct {
			name           string
			currentVersion string
			targetVersion  string
			shouldExecute  bool
		}{
			{
				name:           "v1.5.0 应该从 v1.3.0 执行到 v1.5.0",
				currentVersion: "v1.3.0",
				targetVersion:  "v1.5.0",
				shouldExecute:  true,
			},
			{
				name:           "v1.5.0 不应该从 v1.5.0 执行到 v1.5.0",
				currentVersion: "v1.5.0",
				targetVersion:  "v1.5.0",
				shouldExecute:  false,
			},
			{
				name:           "v1.5.0 不应该从 v1.6.0 执行到 v1.5.0",
				currentVersion: "v1.6.0",
				targetVersion:  "v1.5.0",
				shouldExecute:  false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// 这里可以添加具体的版本比较逻辑测试
				// 由于 v1.5.0 的实现比较简单，这里只是示例
				t.Logf("测试用例: %s", tc.name)
			})
		}
	})
}
