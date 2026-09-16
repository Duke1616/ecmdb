package rule

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRule(t *testing.T) {
	// 读取 JSON 文件
	data, err := os.ReadFile("rule.json")
	assert.NoError(t, err)

	// 直接支持 []byte 解析
	result, err := ParseRules(data)
	assert.NoError(t, err)
	assert.Len(t, result, 5)

	// 同时验证传统 interface{} 解析
	var rules interface{}
	err = json.Unmarshal(data, &rules)
	assert.NoError(t, err)

	result2, err := ParseRules(rules)
	assert.NoError(t, err)
	assert.Len(t, result2, 5)
}

func TestField(t *testing.T) {
	// 读取 JSON 文件
	data, err := os.ReadFile("rule.json")
	assert.NoError(t, err)

	result, err := ParseRules(data)
	assert.NoError(t, err)
	assert.Len(t, result, 5)

	inputData := map[string]interface{}{
		"assets":      []string{"7601628b-3567-472e-af13-4f2c6ad631c4", "b6d7171f-84b1-4778-9f9c-0960cb7b1d42"},
		"environment": "internal",
		"purpose":     "为了荣耀",
		"quantity":    2,
		"role":        1,
	}

	card := GetFields(result, 1, inputData)
	assert.NotEmpty(t, card)

	// 验证选项 Option Label 转换正确（MySQL/MongoDB/Redis）
	assert.Contains(t, card[1].Content, "MongoDB, Redis")
	// 验证环境转换正确
	assert.Contains(t, card[0].Content, "企业内网环境")
	// 验证两列排版自动插入换行 spacer
	assert.Equal(t, "", card[2].Content)

	// 验证无副作用：入参 inputData 未被意外修改
	assert.Len(t, inputData, 5)
}

func TestFilterNotifyHidden(t *testing.T) {
	rules := []map[string]interface{}{
		{
			"type": "input", "field": "password", "title": "密码",
			"notify_hidden": true,
		},
		{
			"type": "input", "field": "token", "title": "令牌",
			"style": map[string]interface{}{"notify_display": "false"},
		},
		{
			"type": "input", "field": "username", "title": "用户名",
			"notify_hidden": false,
		},
	}

	result, err := ParseRules(rules)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.True(t, result[0].NotifyHidden)
	assert.False(t, result[2].NotifyHidden)

	data := map[string]interface{}{
		"password": "secret_password",
		"token":    "secret_token",
		"username": "admin",
	}

	fields := GetFields(result, SystemProvide, data)
	assert.Len(t, fields, 1)
	assert.Contains(t, fields[0].Content, "用户名")
	assert.NotContains(t, fields[0].Content, "密码")
	assert.NotContains(t, fields[0].Content, "令牌")

	// 验证调用 GetFields 之后，原始 map 里的 password 和 token 没有被 delete 篡改
	assert.Contains(t, data, "password")
	assert.Contains(t, data, "token")
	assert.Equal(t, "secret_password", data["password"])
}

func TestFieldOrdering(t *testing.T) {
	// 验证输出字段按照 rules 声明的先后顺序依次排列
	rules := []Rule{
		{Field: "b_field", Title: "第二个字段"},
		{Field: "a_field", Title: "第一个字段"},
	}

	data := map[string]interface{}{
		"a_field": "val_a",
		"b_field": "val_b",
		"c_extra": "val_extra",
	}

	fields := GetFields(rules, SystemProvide, data)
	// 应先出现 b_field，再出现 a_field，最后是 c_extra
	assert.Contains(t, fields[0].Content, "第二个字段")
	assert.Contains(t, fields[1].Content, "第一个字段")
}
