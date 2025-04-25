package domain

import (
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/enotify/notify/feishu/card"
)

type NodeName string

const (
	Start      NodeName = "START"      // 开始节点
	Automation NodeName = "AUTOMATION" // 自动化节点
	User       NodeName = "USER"       // 用户审批节点
)

type StrategyInfo struct {
	NodeName    NodeName          `json:"node_name"`    // 节点名称
	OrderInfo   order.Order       `json:"order_info"`   // 工单提交信息
	WfInfo      workflow.Workflow `json:"wf_info"`      // 流程信息
	InstanceId  int               `json:"instance_id"`  // 实例 id
	CurrentNode *model.Node       `json:"current_node"` // 流程当前节点
}

type Notification struct {
	Receiver string   `json:"receiver"` // 接收者(手机/邮箱/用户ID)
	Template Template `json:"template"` // 发送模版
	Channel  Channel  `json:"channel"`  // 发送渠道
}

type Template struct {
	Name     string       `json:"name"`      // 模版名称
	Title    string       `json:"title"`     // 模版标题
	Fields   []card.Field `json:"fields"`    // 模版字段信息
	Values   []card.Value `json:"values"`    // 模版传递变量
	HideForm bool         `json:"hide_form"` // 隐藏
}
