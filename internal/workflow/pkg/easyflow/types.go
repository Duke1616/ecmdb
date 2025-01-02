package easyflow

import "github.com/Bunny3th/easy-workflow/workflow/model"

type ProcessEngineConvert interface {
	Deploy(workflow Workflow) (int, error)
	Edge(workflow Workflow, tasks []model.Task) ([]string, error)
	GetAutomationProperty(workflow Workflow, nodeId string) (AutomationProperty, error)
}

type UserPropertyType string

func (s UserPropertyType) ToString() string {
	return string(s)
}

const (
	// APPOINT 指定内部人员
	APPOINT UserPropertyType = "appoint"
	// FOUNDER 工单创建人
	FOUNDER UserPropertyType = "founder"
	// TEMPLATE 根据模版字段提取用户
	TEMPLATE UserPropertyType = "template"
)

type Workflow struct {
	Id       int64
	Name     string
	Owner    string
	FlowData LogicFlow
}

type LogicFlow struct {
	Edges []map[string]interface{} `json:"edges"`
	Nodes []map[string]interface{} `json:"nodes"`
}

// Edge 定义线字段
type Edge struct {
	Type         string      `json:"type"`
	SourceNodeId string      `json:"sourceNodeId"`
	TargetNodeId string      `json:"targetNodeId"`
	Properties   interface{} `json:"properties"`
	ID           string      `json:"id"`
}

// Node 节点定义
type Node struct {
	Type       string      `json:"type"`
	Properties interface{} `json:"properties"`
	ID         string      `json:"id"`
}

type EdgeProperty struct {
	Expression string `json:"expression"`
	IsPass     bool   `json:"is_pass"`
}

type UserProperty struct {
	Name          string   `json:"name"`
	Approved      []string `json:"approved"`
	Type          string   `json:"type"`
	TemplateField string   `json:"template_field"`
	IsCosigned    bool     `json:"is_cosigned"`
	IsCC          bool     `json:"is_cc"`
}

type StartProperty struct {
	Name string `json:"name"`
}

type EndProperty struct {
	Name string `json:"name"`
}

type ConditionProperty struct {
	Name string `json:"name"`
}

type AutomationProperty struct {
	Name         string `json:"name"`
	CodebookUid  string `json:"codebook_uid"`
	Tag          string `json:"tag"`
	IsNotify     bool   `json:"is_notify"`
	NotifyMethod int64  `json:"notify_method"`
}
