package easyflow

import (
	"github.com/Bunny3th/easy-workflow/workflow/model"
)

// 节点类型标识，对应前端 LogicFlow DSL 中的 type 字段
const (
	NodeTypeStart     = "start"
	NodeTypeEnd       = "end"
	NodeTypeUser      = "user"
	NodeTypeCondition = "condition"
	NodeTypeParallel  = "parallel"
	NodeTypeInclusion = "inclusion"
	NodeTypeSelective = "selective"
	NodeTypeAuto      = "automation"
	NodeTypeChat      = "chat"
)

// 节点默认名称
const (
	DefaultNameStart     = "Start"
	DefaultNameEnd       = "End"
	DefaultNameUser      = "审批节点"
	DefaultNameParallel  = "并行网关"
	DefaultNameInclusion = "包容网关"
	DefaultNameSelective = "条件并行网关"
)

// 系统内置用户标识
const (
	UserStarter        = "$starter"   // 工单发起人
	AutomationApproval = "automation" // 自动化节点占位用户
	ChatGroupApproval  = "chat_group" // 群通知节点占位用户
	SysAutoUser        = "sys_auto"   // 系统代理节点自动审批用户
)

// SysProxyNodeName 系统代理节点名称
const SysProxyNodeName = "系统代理流转"

// 工作流引擎事件名称
const (
	// EventStart 开始节点触发事件
	EventStart = "EventStart"
	// EventNotify 通用通知事件（节点开始时通知审批人）
	EventNotify = "EventNotify"
	// EventCarbonCopy 抄送事件
	EventCarbonCopy = "EventCarbonCopy"
	// EventAutomation 自动化节点执行事件
	EventAutomation = "EventAutomation"
	// EventChatGroup 群通知节点事件
	EventChatGroup = "EventChatGroup"
	// EventSelectiveGatewaySplit 条件并行网关分裂事件
	EventSelectiveGatewaySplit = "EventSelectiveGatewaySplit"
	// EventRevoke 撤销事件
	EventRevoke = "EventRevoke"

	// EventTaskParallelNodePass 并行网关：分支通过事件
	EventTaskParallelNodePass = "EventTaskParallelNodePass"
	// EventTaskInclusionNodePass 包容网关：分支通过事件
	EventTaskInclusionNodePass = "EventTaskInclusionNodePass"

	// EventConcurrentRejectCleanup 并发拒绝时清理其他分支的事件（condition -> parallel/inclusion/selective）
	EventConcurrentRejectCleanup = "EventConcurrentRejectCleanup"
	// EventInclusionPassCleanup 包容网关通过时清理未完成分支的事件
	EventInclusionPassCleanup = "EventInclusionPassCleanup"
	// EventGatewayConditionReject 网关条件拒绝事件（gateway -> condition 场景下的拒绝）
	EventGatewayConditionReject = "EventGatewayConditionReject"
	// EventUserNodeRejectProxyCleanup 用户节点拒绝时清理代理节点的事件
	EventUserNodeRejectProxyCleanup = "EventUserNodeRejectProxyCleanup"
)

type logicFlow struct {
	converter *DefaultConverter
}

func NewLogicFlowToEngineConvert() Converter {
	return &logicFlow{
		converter: NewDefaultConverterWithHandlers(),
	}
}

func (l *logicFlow) Convert(workflow Workflow) (*model.Process, error) {
	return l.converter.Convert(workflow)
}
