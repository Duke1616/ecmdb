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
)

type OrderEvent struct {
	Id         int64                  `json:"id"`
	WorkflowId int64                  `json:"workflow_id"`
	Provide    Provide                `json:"provide"`
	Data       map[string]interface{} `json:"data"`
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
