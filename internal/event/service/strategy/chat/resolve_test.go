package chat

import (
	"testing"

	"github.com/Duke1616/ecmdb/internal/event/service/strategy"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/stretchr/testify/assert"
)

func TestResolveDynamicString(t *testing.T) {
	n := &Notification{} // 我们只需要测试纯方法，不需要 Mock Service

	testCases := []struct {
		name       string
		ruleData   string
		defaultVal string
		info       strategy.Info
		data       *chatContext
		expected   string
	}{
		{
			name:       "标准变量全部替换",
			ruleData:   "工单[{{ticket_id}}]由{{creator}}发送",
			defaultVal: "默认标题",
			info: strategy.Info{
				FlowContext: strategy.FlowContext{
					Order: order.Order{Id: 9988},
				},
			},
			data: &chatContext{
				NotificationData: &strategy.NotificationData{
					StartUser: user.User{DisplayName: "张三"},
				},
			},
			expected: "工单[9988]由张三发送",
		},
		{
			name:       "包含自定义表单字段替换",
			ruleData:   "项目: {{field.project_name}}, 环境: {{field.env}}",
			defaultVal: "默认",
			info: strategy.Info{
				FlowContext: strategy.FlowContext{
					Order: order.Order{
						Data: map[string]interface{}{
							"project_name": "ECMDB",
							"env":          "PROD",
						},
					},
				},
			},
			data: &chatContext{
				NotificationData: &strategy.NotificationData{},
			},
			expected: "项目: ECMDB, 环境: PROD",
		},
		{
			name:       "部分变量不存在时原样保留",
			ruleData:   "发起人: {{creator}}, 应用: {{field.app}}, 缺失: {{not_exist}}",
			defaultVal: "默认",
			info: strategy.Info{
				FlowContext: strategy.FlowContext{
					Order: order.Order{
						Data: map[string]interface{}{"app": "Redis"},
					},
				},
			},
			data: &chatContext{
				NotificationData: &strategy.NotificationData{
					StartUser: user.User{DisplayName: "李四"},
				},
			},
			expected: "发起人: 李四, 应用: Redis, 缺失: {{not_exist}}",
		},
		{
			name:       "规则为空使用默认值",
			ruleData:   "",
			defaultVal: "兜底群名称",
			info:       strategy.Info{},
			data: &chatContext{
				NotificationData: &strategy.NotificationData{},
			},
			expected: "兜底群名称",
		},
		{
			name:       "混合复杂类型字段(转字符串)",
			ruleData:   "端口: {{field.port}}, 状态: {{field.is_ok}}",
			defaultVal: "默认",
			info: strategy.Info{
				FlowContext: strategy.FlowContext{
					Order: order.Order{
						Data: map[string]interface{}{
							"port":  3306,
							"is_ok": true,
						},
					},
				},
			},
			data: &chatContext{
				NotificationData: &strategy.NotificationData{},
			},
			expected: "端口: 3306, 状态: true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := n.resolveDynamicString(tc.ruleData, tc.defaultVal, tc.info, tc.data)
			assert.Equal(t, tc.expected, result)
		})
	}
}
