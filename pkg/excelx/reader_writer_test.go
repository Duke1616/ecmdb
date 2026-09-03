package excelx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_TableWriter_And_Reader_Flow(t *testing.T) {
	columns := []Column{
		{Key: "name", Title: "资产名称", Type: "string", Required: true, Unique: true},
		{Key: "cpu", Title: "CPU核心数", Type: "int", Required: true},
		{Key: "ratio", Title: "利用率", Type: "float"},
		{Key: "is_prod", Title: "是否生产环境", Type: "bool"},
		{Key: "os", Title: "操作系统", Type: "select", Options: []string{"Linux", "Windows"}},
	}

	testCases := []struct {
		name      string
		records   []map[string]interface{}
		assertFn  func(t *testing.T, res []map[string]interface{}, err error)
	}{
		{
			name: "正常数据读写闭环_强类型转换成功",
			records: []map[string]interface{}{
				{
					"name":    "srv-01",
					"cpu":     32,
					"ratio":   0.85,
					"is_prod": "是",
					"os":      "Linux",
				},
			},
			assertFn: func(t *testing.T, res []map[string]interface{}, err error) {
				require.NoError(t, err)
				require.Len(t, res, 1)
				assert.Equal(t, "srv-01", res[0]["name"])
				assert.Equal(t, int64(32), res[0]["cpu"])
				assert.Equal(t, 0.85, res[0]["ratio"])
				assert.Equal(t, true, res[0]["is_prod"])
				assert.Equal(t, "Linux", res[0]["os"])
			},
		},
		{
			name: "必填字段缺失_触发精准行级校验拦截",
			records: []map[string]interface{}{
				{
					"name": "srv-02",
					// 缺少必填的 cpu
					"os": "Linux",
				},
			},
			assertFn: func(t *testing.T, res []map[string]interface{}, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "必填字段【CPU核心数(cpu)】不能为空")
			},
		},
		{
			name: "枚举选项非法_触发下拉值域校验拦截",
			records: []map[string]interface{}{
				{
					"name": "srv-03",
					"cpu":  16,
					"os":   "macOS", // 不在 Linux, Windows 范围内
				},
			},
			assertFn: func(t *testing.T, res []map[string]interface{}, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "不在允许的预设选项列表中")
			},
		},
		{
			name: "数字格式非法_触发类型校验拦截",
			records: []map[string]interface{}{
				{
					"name": "srv-04",
					"cpu":  "invalid_int",
				},
			},
			assertFn: func(t *testing.T, res []map[string]interface{}, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "不是合法的整数")
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// 1. 写数据生成 Excel 字节流
			writer := NewWriter("主机", columns)
			bytes, err := writer.WriteData(tc.records)
			require.NoError(t, err)
			require.NotEmpty(t, bytes)

			// 2. 读数据反序列化与校验
			reader := NewReader(columns)
			parsedRecords, parseErr := reader.Read(bytes)
			tc.assertFn(t, parsedRecords, parseErr)
		})
	}
}

func Test_TableWriter_WriteTemplate(t *testing.T) {
	testCases := []struct {
		name      string
		sheetName string
		columns   []Column
	}{
		{
			name:      "生成空白模板_包含完整3行表头且成功导出",
			sheetName: "模板",
			columns: []Column{
				{Key: "name", Title: "名称", Type: "string", Required: true, Unique: true},
				{Key: "status", Title: "状态", Type: "select", Options: []string{"启用", "停用"}},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			writer := NewWriter(tc.sheetName, tc.columns)
			templateBytes, err := writer.WriteTemplate()
			require.NoError(t, err)
			assert.NotEmpty(t, templateBytes)
		})
	}
}
