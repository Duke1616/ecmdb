// Copyright 2023 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package event

const (
	// WechatOrderEventName 接收企业微信 OA 事件
	WechatOrderEventName = "wechat_order_events"
	// CreateProcessEventName 创建流程事件
	CreateProcessEventName = "create_process_events"
	// OrderStatusModifyEventName 修改状态事件
	OrderStatusModifyEventName = "order_status_modify_events"
	// FeishuCallbackEventName 飞书回调事件
	FeishuCallbackEventName = "feishu_callback_events"
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

type FeishuCallback struct {
	Action       string `json:"action"`
	MessageId    string `json:"message_id"`
	FeishuUserId string `json:"feishu_user_id"`
	TaskId       string `json:"task_id"`
	Comment      string `json:"comment"`
	OrderId      string `json:"order_id"`
}
