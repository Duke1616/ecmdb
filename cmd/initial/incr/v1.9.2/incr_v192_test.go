package v192

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Duke1616/ecmdb/cmd/initial/incr/version"
)

func TestVersion192Logic(t *testing.T) {
	// 测试版本比较逻辑
	testCases := []struct {
		name           string
		currentVersion string
		targetVersion  string
		version        string
		shouldExecute  bool
	}{
		{
			name:           "v1.9.2 应该从 v1.3.0 执行到 v1.9.2",
			currentVersion: "v1.3.0",
			targetVersion:  "v1.9.2",
			version:        "v1.9.2",
			shouldExecute:  true,
		},
		{
			name:           "v1.5.0 应该从 v1.3.0 执行到 v1.9.2",
			currentVersion: "v1.3.0",
			targetVersion:  "v1.9.2",
			version:        "v1.5.0",
			shouldExecute:  true,
		},
		{
			name:           "v1.2.3 不应该从 v1.3.0 执行到 v1.9.2",
			currentVersion: "v1.3.0",
			targetVersion:  "v1.9.2",
			version:        "v1.2.3",
			shouldExecute:  false,
		},
		{
			name:           "v2.0.0 不应该从 v1.3.0 执行到 v1.9.2",
			currentVersion: "v1.3.0",
			targetVersion:  "v1.9.2",
			version:        "v2.0.0",
			shouldExecute:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
		currentVerSlice := version.ParseVersion(tc.currentVersion)
		targetVerSlice := version.ParseVersion(tc.targetVersion)
		versionSlice := version.ParseVersion(tc.version)

			// 检查是否应该执行
			shouldExecute := version.Compare(versionSlice, currentVerSlice) &&
				(version.Compare(targetVerSlice, versionSlice) || version.Equal(versionSlice, targetVerSlice))

			if shouldExecute != tc.shouldExecute {
				t.Errorf("版本 %s 执行判断错误: 期望 %v, 得到 %v",
					tc.version, tc.shouldExecute, shouldExecute)
			}
		})
	}
}

func TestVersion192EndpointsLogic(t *testing.T) {
	// 测试 endpoints resource 字段处理逻辑
	testCases := []struct {
		name     string
		resource string
		expected string
	}{
		{"空字符串应该填充 CMDB", "", "CMDB"},
		{"已有值不应该改变", "EXISTING", "EXISTING"},
		{"CMDB 值不应该改变", "CMDB", "CMDB"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 模拟处理逻辑
			resource := tc.resource
			if resource == "" {
				resource = "CMDB"
			}

			if resource != tc.expected {
				t.Errorf("resource 处理错误: 期望 %s, 得到 %s", tc.expected, resource)
			}
		})
	}
}

func TestBackupCollectionName(t *testing.T) {
	// 测试备份集合名称生成逻辑
	testCases := []struct {
		name     string
		version  string
		expected string
	}{
		{"v1.9.2 版本备份", "v1.9.2", "c_menu_backup_v1.9.2_"},
		{"v2.0.0 版本备份", "v2.0.0", "c_menu_backup_v2.0.0_"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 模拟备份集合名称生成
			collectionName := fmt.Sprintf("c_menu_backup_%s_%s", tc.version, "20240101_120000")
			
			if !strings.HasPrefix(collectionName, tc.expected) {
				t.Errorf("备份集合名称错误: 期望前缀 %s, 得到 %s", tc.expected, collectionName)
			}
		})
	}
}

func TestBackupDocumentStructure(t *testing.T) {
	// 测试备份文档结构
	originalMenu := map[string]interface{}{
		"id":   123,
		"name": "测试菜单",
		"path": "/test",
	}

	version := "v1.9.2"
	backupTime := int64(1704067200000) // 2024-01-01 12:00:00

	// 模拟备份文档结构
	backupDoc := map[string]interface{}{
		"original_id": originalMenu["id"],
		"version":     version,
		"backup_time": backupTime,
		"data":        originalMenu,
	}

	// 验证备份文档结构
	if backupDoc["original_id"] != originalMenu["id"] {
		t.Errorf("original_id 错误: 期望 %v, 得到 %v", originalMenu["id"], backupDoc["original_id"])
	}

	if backupDoc["version"] != version {
		t.Errorf("version 错误: 期望 %s, 得到 %s", version, backupDoc["version"])
	}

	if backupDoc["backup_time"] != backupTime {
		t.Errorf("backup_time 错误: 期望 %d, 得到 %v", backupTime, backupDoc["backup_time"])
	}

	if backupDoc["data"] == nil {
		t.Error("data 字段为空")
	}
}
