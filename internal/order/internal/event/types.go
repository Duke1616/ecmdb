package event

import (
	"fmt"
	"strconv"
)

const (
	// WechatOrderEventName 接收企业微信 OA 事件
	WechatOrderEventName = "wechat_order_events"
	// CreateProcessEventName 创建流程事件
	CreateProcessEventName = "create_process_events"
	// OrderStatusModifyEventName 修改状态事件
	OrderStatusModifyEventName = "order_status_modify_events"
	// LarkCallbackEventName 飞书回调事件
	LarkCallbackEventName = "lark_callback_events"
)

type OrderEvent struct {
	Id         int64                  `json:"id"`
	WorkflowId int64                  `json:"workflow_id"`
	Provide    Provide                `json:"provide"`
	Data       map[string]interface{} `json:"data"`
	// 流程引擎使用的变量, 根据这样可以定制express判断公式
	Variables string `json:"variables"`
}

type Variables struct {
	Key   string
	Value any
}

type Status uint8

func (s Status) ToUint8() uint8 {
	return uint8(s)
}

type OrderStatusModifyEvent struct {
	ProcessInstanceId int    `json:"process_instance_id"`
	Status            Status `json:"status"`
}

type Provide uint8

func (s Provide) ToUint8() uint8 {
	return uint8(s)
}

const (
	// SYSTEM 本系统
	SYSTEM Provide = 1
	// WECHAT 企业微信
	WECHAT Provide = 2
)

type Action string

const (
	Pass     Action = "pass"
	Reject   Action = "reject"
	Progress Action = "progress"
	Revoke   Action = "revoke"
)

type LarkCallback struct {
	Action    Action                 `json:"action"`
	MessageId string                 `json:"message_id"`
	UserId    string                 `json:"user_id"`
	OpenId    string                 `json:"open_id"`
	FormValue map[string]interface{} `json:"form_value"`
	Value     map[string]interface{} `json:"value"`
}

func (l *LarkCallback) GetMessageId() string {
	return l.MessageId
}

func (l *LarkCallback) GetOrderId() string {
	if v, ok := l.Value["order_id"].(string); ok {
		return v
	}
	return ""
}

func (l *LarkCallback) GetOrderIdInt() (int64, error) {
	val := l.GetOrderId()
	if val == "" {
		return 0, fmt.Errorf("order_id is empty")
	}
	return strconv.ParseInt(val, 10, 64)
}

func (l *LarkCallback) GetTaskId() string {
	if v, ok := l.Value["task_id"].(string); ok {
		return v
	}
	return ""
}

func (l *LarkCallback) GetTaskIdInt() (int, error) {
	val := l.GetTaskId()
	if val == "" {
		return 0, fmt.Errorf("task_id is empty")
	}
	return strconv.Atoi(val)
}

func (l *LarkCallback) GetAction() Action {
	if l.Action != "" {
		return l.Action

	}

	if v, ok := l.Value["action"].(string); ok {
		return Action(v)
	}

	return ""
}

func (l *LarkCallback) GetComment() string {
	if v, ok := l.FormValue["comment"].(string); ok {
		if v == "" {
			return "无"
		}
		return v
	}
	return "无"
}

func (l *LarkCallback) GetUserId() string {
	return l.UserId
}

func (l *LarkCallback) GetOpenId() string {
	return l.OpenId
}
