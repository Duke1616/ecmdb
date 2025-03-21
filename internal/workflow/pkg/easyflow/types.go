package easyflow

import "github.com/Bunny3th/easy-workflow/workflow/model"

type ProcessEngineConvert interface {
	Deploy(workflow Workflow) (int, error)
	Edge(workflow Workflow, tasks []model.Task) ([]string, error)
	GetAutomationProperty(workflow Workflow, nodeId string) (AutomationProperty, error)
}

type Rule string

func (s Rule) ToString() string {
	return string(s)
}

const (
	// APPOINT 指定内部人员
	APPOINT Rule = "appoint"
	// FOUNDER 工单创建人
	FOUNDER Rule = "founder"
	// TEMPLATE 根据模版字段提取用户
	TEMPLATE Rule = "template"
	// LEADER 部门领导
	LEADER Rule = "leaders"
	// MAIN_LEADER 分管领导
	MAIN_LEADER Rule = "main_leader"
)

type ExecMethod string

func (s ExecMethod) ToString() string {
	return string(s)
}

const (
	// EXEC_TEMPLATE 根据模版
	EXEC_TEMPLATE ExecMethod = "template"
	// HAND 手动方式
	HAND ExecMethod = "hand"
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
	Type         string                   `json:"type"`
	SourceNodeId string                   `json:"sourceNodeId"`
	TargetNodeId string                   `json:"targetNodeId"`
	Properties   interface{}              `json:"properties"`
	ID           string                   `json:"id"`
	StartPoint   map[string]interface{}   `json:"startPoint"`
	EndPoint     map[string]interface{}   `json:"endPoint"`
	PointsList   []map[string]interface{} `json:"pointsList"`
	Text         map[string]interface{}   `json:"text"`
}

// Node 节点定义
type Node struct {
	Type       string      `json:"type"`
	Properties interface{} `json:"properties"`
	ID         string      `json:"id"`
}

type EdgeProperty struct {
	Name       string `json:"name"`
	Expression string `json:"expression"` // 表达式
	IsPass     bool   `json:"is_pass"`    // 连线是否通过、为了绘制流程图走向使用
}

type UserProperty struct {
	Name          string   `json:"name"`           // 节点名称
	Approved      []string `json:"approved"`       // 审批人、抄送人
	Rule          Rule     `json:"rule"`           // 匹配策略
	TemplateField string   `json:"template_field"` // 模版字段
	IsCosigned    bool     `json:"is_cosigned"`    // 是否会签
	IsCC          bool     `json:"is_cc"`          // 是否抄送
}

type StartProperty struct {
	Name     string `json:"name"`
	IsNotify bool   `json:"is_notify"` // 是否开启开始节点消息通知
}

type EndProperty struct {
	Name string `json:"name"`
}

type ConditionProperty struct {
	Name string `json:"name"`
}

type AutomationProperty struct {
	Name          string  `json:"name"`
	CodebookUid   string  `json:"codebook_uid"`   // 代码库UID
	Tag           string  `json:"tag"`            // runner tags
	IsNotify      bool    `json:"is_notify"`      // 是否开始消息通知
	Unit          uint8   `json:"unit"`           // 定时执行：单位
	Quantity      int64   `json:"quantity"`       // 定时执行：数量
	ExecMethod    string  `json:"exec_method"`    // 执行方式, template 模版获取，hand 手动指定
	TemplateId    int64   `json:"template_id"`    // 模版ID
	TemplateField string  `json:"template_field"` // 模版字段
	IsTiming      bool    `json:"is_timing"`      // 是否开始定时执行
	NotifyMethod  []int64 `json:"notify_method"`  // 消息通知模式
}
