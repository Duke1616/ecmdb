package easyflow

import (
	"fmt"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/mitchellh/mapstructure"
)

// Converter 工作流转换器接口
type Converter interface {
	// Convert 将工作流 DSL 转换为流程引擎模型
	Convert(workflow Workflow) (*model.Process, error)
}

// INodeHandler 节点解析器接口
type INodeHandler interface {
	Type() string
	Handle(ctx *Context, node Node) ([]model.Node, error)
}

// Context 转换过程的上下文
type Context struct {
	Workflow     Workflow
	NodesMap     map[string]Node     // ID -> DSL Node
	EdgesMap     map[string][]Edge   // SourceNodeId -> []Edge
	PrevNodesMap map[string][]string // TargetNodeId -> []SourceNodeId
	OutputNodes  []model.Node        // 转换生成的最终节点集
	ProxyNodes   map[string]struct{} // 已创建的代理节点 ID 集合，防止重复
}

// GetPrevNodeIDs 获取目标节点的上级节点 ID
func (ctx *Context) GetPrevNodeIDs(nodeID string) []string {
	return ctx.PrevNodesMap[nodeID]
}

// GetTargetEdges 获取指定节点的出边
func (ctx *Context) GetTargetEdges(nodeID string) []Edge {
	return ctx.EdgesMap[nodeID]
}

// GetNodeInfo 获取原始节点信息
func (ctx *Context) GetNodeInfo(nodeID string) Node {
	return ctx.NodesMap[nodeID]
}

// AddOutputNode 添加生成后的 Engine 节点
func (ctx *Context) AddOutputNode(n ...model.Node) {
	ctx.OutputNodes = append(ctx.OutputNodes, n...)
}

// GetOrGenerateProxyID 获取或生成代理节点 ID
// 当目标节点是网关类型时，创建并返回代理节点 ID
// 否则返回目标节点 ID
func (ctx *Context) GetOrGenerateProxyID(srcID, targetID string) string {
	if ctx.ProxyNodes == nil {
		ctx.ProxyNodes = make(map[string]struct{})
	}

	targetNode := ctx.GetNodeInfo(targetID)
	// 只有目标是特定网关类型时才需要代理
	if targetNode.Type == "parallel" || targetNode.Type == "inclusion" || targetNode.Type == "selective" {
		proxyID := fmt.Sprintf("proxy_%s_%s", srcID, targetID)
		if _, ok := ctx.ProxyNodes[proxyID]; !ok {
			ctx.createProxyNode(proxyID, srcID, targetID)
		}
		return proxyID
	}
	return targetID
}

// GetOrGenerateProxyForGateway 当源节点是条件网关连接到并行/包容/条件并行网关时，获取或生成代理节点 ID
// 用于处理条件网关到其他网关的连接
func (ctx *Context) GetOrGenerateProxyForGateway(conditionGatewayID, targetID string) string {
	if ctx.ProxyNodes == nil {
		ctx.ProxyNodes = make(map[string]struct{})
	}

	conditionNode := ctx.GetNodeInfo(conditionGatewayID)
	// 只有源节点是条件网关时才需要代理
	if conditionNode.Type == "condition" {
		proxyID := fmt.Sprintf("proxy_%s_%s", conditionGatewayID, targetID)
		if _, ok := ctx.ProxyNodes[proxyID]; !ok {
			ctx.createProxyNodeForGateway(proxyID, conditionGatewayID, targetID)
		}
		return proxyID
	}
	return conditionGatewayID
}

func (ctx *Context) createProxyNode(proxyID, srcID, targetID string) {
	if ctx.ProxyNodes == nil {
		ctx.ProxyNodes = make(map[string]struct{})
	}
	ctx.ProxyNodes[proxyID] = struct{}{}

	targetNode := ctx.GetNodeInfo(targetID)
	var passEvents []string
	switch targetNode.Type {
	case "parallel":
		passEvents = []string{"EventTaskParallelNodePass"}
	case "inclusion":
		passEvents = []string{"EventTaskInclusionNodePass"}
	case "selective":
		passEvents = []string{"EventTaskSelectiveNodePass"}
	}

	n := model.Node{
		NodeID:           proxyID,
		NodeName:         SysProxyNodeName,
		NodeType:         1,
		PrevNodeIDs:      []string{srcID},
		UserIDs:          []string{SysAutoUser},
		NodeStartEvents:  []string{"EventNotify"},
		NodeEndEvents:    []string{},
		TaskFinishEvents: passEvents,
	}
	ctx.AddOutputNode(n)
}

func (ctx *Context) createProxyNodeForGateway(proxyID, conditionGatewayID, targetID string) {
	if ctx.ProxyNodes == nil {
		ctx.ProxyNodes = make(map[string]struct{})
	}
	ctx.ProxyNodes[proxyID] = struct{}{}

	targetNode := ctx.GetNodeInfo(targetID)
	var passEvents []string
	switch targetNode.Type {
	case "parallel":
		passEvents = []string{"EventTaskParallelNodePass"}
	case "inclusion":
		passEvents = []string{"EventTaskInclusionNodePass"}
	case "selective":
		passEvents = []string{"EventTaskSelectiveNodePass"}
	}

	n := model.Node{
		NodeID:           proxyID,
		NodeName:         SysProxyNodeName,
		NodeType:         1,
		PrevNodeIDs:      []string{conditionGatewayID},
		UserIDs:          []string{SysAutoUser},
		NodeStartEvents:  []string{"EventNotify"},
		NodeEndEvents:    []string{},
		TaskFinishEvents: passEvents,
	}
	ctx.AddOutputNode(n)
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
	// ON_CALL 值班排班人员
	ON_CALL Rule = "on_call"
	// TEAM 团队人员
	TEAM Rule = "team"
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

type Node struct {
	Type       string      `json:"type"`
	Properties interface{} `json:"properties"`
	ID         string      `json:"id"`
}

func ParseNodes(raw any) ([]Node, error) {
	var nodes []Node

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &nodes,
		TagName: "json",
	})
	if err != nil {
		return nil, err
	}

	if err = decoder.Decode(raw); err != nil {
		return nil, err
	}

	return nodes, nil
}

type EdgeProperty struct {
	Name       string `json:"name"`
	Expression string `json:"expression"` // 表达式
	IsPass     bool   `json:"is_pass"`    // 连线是否通过、为了绘制流程图走向使用
}

type FieldType string

func (f FieldType) ToString() string {
	return string(f)
}

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
	// FieldTips 提示选项
	FieldTips FieldType = "tips"
)

type Option struct {
	Label string `json:"label"` // 选项显示名
	Value string `json:"value"` // 选项值
}

type Field struct {
	Name     string            `json:"name"`     // 表单字段显示名
	Key      string            `json:"key"`      // 表单字段键名（对应 Order Data Key）
	Type     FieldType         `json:"type"`     // 字段类型：input, textarea, date, number...
	Required bool              `json:"required"` // 是否必填
	Options  []Option          `json:"options"`  // 选项列表（用于 select 等）
	Props    map[string]string `json:"props"`    // 额外组件属性（如 placeholder）
	Merge    bool              `json:"merge"`    // 如果 Merge 则在后续审批节点进行推送展示
	Validate string            `json:"validate"` // 数据校验
	Hidden   bool              `json:"hidden"`   // 字段需要，但是不展示，由系统补充
	Value    string            `json:"value"`    // 数据值
	ReadOnly bool              `json:"readonly"` // 只读字段，比如提示用户时候使用
}

// Assignee 审批人员分配规则配置
type Assignee struct {
	Rule   Rule     `json:"rule"`   // 匹配策略
	Values []string `json:"values"` // 规则的目标值列表（使用 string 兼容更多实体标识）
}

type UserProperty struct {
	Name          string     `json:"name"`           // 节点名称
	Approved      []string   `json:"approved"`       // 审批人、抄送人
	Rule          Rule       `json:"rule"`           // 匹配策略 (兼容历史数据)
	Type          Rule       `json:"type"`           // 匹配策略 (兼容历史数据
	TemplateField string     `json:"template_field"` // 模版字段
	Assignees     []Assignee `json:"assignees"`      // 新模式字段，支持配置多条分配规则
	IsCosigned    bool       `json:"is_cosigned"`    // 是否会签
	IsCC          bool       `json:"is_cc"`          // 是否抄送
	Fields        []Field    `json:"fields"`         // 表单字段配置
}

func (u *UserProperty) getRule() Rule {
	if u.Rule != "" {
		return u.Rule
	}
	return u.Type
}

// NormalizeAssignees 统一格式化获取人员分配规则，屏蔽新老版本数据差异
func (u *UserProperty) NormalizeAssignees() []Assignee {
	// 默认将使用新版本模式
	if len(u.Assignees) > 0 {
		return u.Assignees
	}

	// 兼容老版本情况
	switch u.getRule() {
	case TEMPLATE:
		return []Assignee{
			{
				Rule:   u.Rule,
				Values: []string{u.TemplateField},
			},
		}
	default:
		return []Assignee{
			{
				Rule:   u.Rule,
				Values: u.Approved,
			},
		}
	}
}

type StartProperty struct {
	Name     string `json:"name"`
	IsNotify bool   `json:"is_notify"` // 是否开启开始节点消息通知
}

type ChatGroupMode string

const (
	ChatGroupUseExisting ChatGroupMode = "existing"
	ChatGroupCreate      ChatGroupMode = "create"
)

type OutputMode string

const (
	OutputTicketData OutputMode = "ticket_data" // 工单提交信息
	OutputAutoTask   OutputMode = "auto_task"   // 自动化任务返回结果
	OutputUserInput  OutputMode = "user_input"  // 用户节点提交信息
)

// ChatGroupProperty 群通知节点属性
// 该节点为纯广播型，发送完成后自动推进流程，无需等待任何操作
type ChatGroupProperty struct {
	Name         string           `json:"name"`                     // 节点名称
	Title        string           `json:"title"`                    // 消息卡片标题
	Mode         ChatGroupMode    `json:"mode"`                     // existing / create
	ChatGroupIDs []int64          `json:"chat_group_ids,omitempty"` // existing 模式, 自动匹配所属 team 内部的所有人
	Create       *CreateChatGroup `json:"create,omitempty"`         // create 模式，新建一个群组，全局不绑定任何 Team，或者默认 Team
	Assignees    []Assignee       `json:"assignees"`                // 成员规则
	OutputMode   []OutputMode     `json:"is_auto"`                  // 支持的返回数据
}

type CreateChatGroup struct {
	Name    string `json:"name"`    // 创建群名称
	Channel string `json:"channel"` // 通知渠道
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
