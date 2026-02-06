package domain

import (
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/workflow"
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

type FieldType string

const (
	// FieldInput 单行文本
	FieldInput FieldType = "input"
	// FieldTextarea 多行文本
	FieldTextarea FieldType = "textarea"
	// FieldNumber 数字
	FieldNumber FieldType = "number"
	// FieldDate 日期
	FieldDate FieldType = "date"
	// FieldSelect 下拉选择
	FieldSelect FieldType = "select"
	// FieldMultiSelect 多项选择
	FieldMultiSelect FieldType = "multi_select"
)

type InputOption struct {
	Label string `json:"label"` // 选项显示名
	Value string `json:"value"` // 选项值
}

type InputField struct {
	Name     string            `json:"name"`     // 表单字段显示名
	Key      string            `json:"key"`      // 表单字段键名（对应 Order Data Key）
	Type     FieldType         `json:"type"`     // 字段类型：input, textarea, date, number...
	Required bool              `json:"required"` // 是否必填
	Options  []InputOption     `json:"options"`  // 选项列表（用于 select 等）
	Props    map[string]string `json:"props"`    // 额外组件属性（如 placeholder）
}

type Template struct {
	Name        string       `json:"name"`         // 模版名称
	Title       string       `json:"title"`        // 模版标题
	Fields      []Field      `json:"fields"`       // 模版字段信息
	Values      []Value      `json:"values"`       // 模版传递变量
	InputFields []InputField `json:"input_fields"` // 录入的字段
	HideForm    bool         `json:"hide_form"`    // 隐藏
}

type Field struct {
	IsShort bool   `json:"is_short"`
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

type Value struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}
