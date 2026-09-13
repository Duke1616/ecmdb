package codec

import (
	"reflect"
	"testing"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/stretchr/testify/assert"
)

func TestParsePluginTag(t *testing.T) {
	type sample struct {
		Ignored     string `plugin:"-"`
		IP          string `plugin:"ip,required"`
		Port        int    `plugin:"port,label=SSH端口,field=ssh_port,default=22"`
		DefaultOnly string `plugin:"required,default=hello"`
		Gateways    []any  `plugin:"gateways,model=gateway,out=default"`
		SubGateways []any  `plugin:"sub_gateways,model=sub_gateway,in=run"`
		Embedded    string `plugin:",label=嵌入模型,group=安全凭据"`
		NoTag       string `json:"custom_field"`
		OnlyGoName  string
	}

	st := reflect.TypeOf(sample{})

	testCases := []struct {
		name      string
		fieldName string
		assertTag func(t *testing.T, tag pluginTag)
	}{
		{
			name:      "跳过忽略字段",
			fieldName: "Ignored",
			assertTag: func(t *testing.T, tag pluginTag) {
				assert.True(t, tag.skip)
			},
		},
		{
			name:      "基础属性与必填",
			fieldName: "IP",
			assertTag: func(t *testing.T, tag pluginTag) {
				assert.Equal(t, "ip", tag.key)
				assert.True(t, tag.required)
				assert.Equal(t, "ip", tag.CMDBUID())
				assert.Equal(t, "ip", tag.DisplayName())
			},
		},
		{
			name:      "丰富属性标签解析 (field/label/default)",
			fieldName: "Port",
			assertTag: func(t *testing.T, tag pluginTag) {
				assert.Equal(t, "port", tag.key)
				assert.Equal(t, "ssh_port", tag.CMDBUID())
				assert.Equal(t, "SSH端口", tag.DisplayName())
				assert.Equal(t, "22", tag.defaultValue)
			},
		},
		{
			name:      "保留关键字首项不误作为字段名",
			fieldName: "DefaultOnly",
			assertTag: func(t *testing.T, tag pluginTag) {
				assert.Equal(t, "defaultOnly", tag.key)
				assert.True(t, tag.required)
				assert.Equal(t, "hello", tag.defaultValue)
			},
		},
		{
			name:      "模型出向关联快捷关系",
			fieldName: "Gateways",
			assertTag: func(t *testing.T, tag pluginTag) {
				assert.Equal(t, "gateways", tag.key)
				assert.Equal(t, "gateway", tag.ModelUID())
				assert.Equal(t, "default", tag.relationType)
				assert.Equal(t, types.DirectionToTarget, tag.direction)
			},
		},
		{
			name:      "模型入向关联快捷关系",
			fieldName: "SubGateways",
			assertTag: func(t *testing.T, tag pluginTag) {
				assert.Equal(t, "sub_gateways", tag.key)
				assert.Equal(t, "sub_gateway", tag.ModelUID())
				assert.Equal(t, "run", tag.relationType)
				assert.Equal(t, types.DirectionToSource, tag.direction)
			},
		},
		{
			name:      "匿名嵌入模型标签",
			fieldName: "Embedded",
			assertTag: func(t *testing.T, tag pluginTag) {
				assert.Equal(t, "嵌入模型", tag.DisplayName())
				assert.Equal(t, "安全凭据", tag.group)
			},
		},
		{
			name:      "无标签回退至 json tag",
			fieldName: "NoTag",
			assertTag: func(t *testing.T, tag pluginTag) {
				assert.Equal(t, "custom_field", tag.key)
			},
		},
		{
			name:      "无标签且无 json 回退至 Go 字段名首字母小写",
			fieldName: "OnlyGoName",
			assertTag: func(t *testing.T, tag pluginTag) {
				assert.Equal(t, "onlyGoName", tag.key)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			field, ok := st.FieldByName(tc.fieldName)
			assert.True(t, ok)
			tag := parsePluginTag(field)
			tc.assertTag(t, tag)
		})
	}
}
