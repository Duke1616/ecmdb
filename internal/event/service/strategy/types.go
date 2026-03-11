package strategy

import (
	"context"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
)

type NodeName string

const (
	Start      NodeName = "START"       // 开始节点
	Automation NodeName = "AUTOMATION"  // 自动化节点
	User       NodeName = "USER"        // 用户审批节点
	CarbonCopy NodeName = "CARBON_COPY" // 抄送节点 (Carbon Copy)
	ChatGroup  NodeName = "CHAT_GROUP"  // 群通知节点
)

const (
	// LarkTemplateApprovalName 正常审批通知
	LarkTemplateApprovalName = workflow.NotifyTypeApproval
	// LarkTemplateApprovalRevokeName 带有撤销的审批通知
	LarkTemplateApprovalRevokeName = workflow.NotifyTypeRevoke
	// LarkTemplateCC 抒送通知（用于抒送节点）
	LarkTemplateCC = workflow.NotifyTypeCC
	// LarkTemplateChatGroup 群通知（用于群组节点，支持小标题 + hr 分隔线）
	LarkTemplateChatGroup = workflow.NotifyTypeChat
)

type FlowContext struct {
	InstID      int                // 流程实例 ID
	Order       order.Order        // 工单实体
	Workflow    workflow.Workflow  // 流程定义快照
	Instance    engineSvc.Instance // 引擎实例状态
	CurrentNode *model.Node        // 当前触发事件的节点
	Nodes       []easyflow.Node    // 解析后的流程节点
}

type Info struct {
	NodeName    NodeName             `json:"node_name"` // 节点名称
	Channel     notification.Channel `json:"channel"`   // 通知渠道
	FlowContext                      // 嵌入流程上下文
}

// SendStrategy 针对不同节点的策略接口
type SendStrategy interface {
	Send(ctx context.Context, info Info) (notification.NotificationResponse, error)
}

type Dispatcher struct {
	userStrategy       SendStrategy
	autoStrategy       SendStrategy
	startStrategy      SendStrategy
	chatStrategy       SendStrategy
	carbonCopyStrategy SendStrategy
	base               BaseStrategy
}

func NewDispatcher(user *UserNotification, auto *AutomationNotification,
	start *StartNotification, chat *ChatNotification, carbonCopy *CarbonCopyNotification, base BaseStrategy) *Dispatcher {
	return &Dispatcher{
		userStrategy:       user,
		autoStrategy:       auto,
		startStrategy:      start,
		chatStrategy:       chat,
		carbonCopyStrategy: carbonCopy,
		base:               base,
	}
}

func (d *Dispatcher) Send(ctx context.Context, info Info) (notification.NotificationResponse, error) {
	strategy := d.selectStrategy(info)

	// 1. 预解析流程节点，避免策略内重复解析
	if nodes, err := d.base.UnmarshalNodes(info.Workflow); err == nil {
		info.Nodes = nodes
	}

	// 2. 分发器根据流程属性注入通知渠道类型
	if info.Channel == "" {
		info.Channel = d.base.GetChannel(info.Workflow)
	}

	return strategy.Send(ctx, info)
}

func (d *Dispatcher) selectStrategy(not Info) SendStrategy {
	switch not.NodeName {
	case Start:
		return d.startStrategy
	case Automation:
		return d.autoStrategy
	case User:
		return d.userStrategy
	case CarbonCopy:
		return d.carbonCopyStrategy
	case ChatGroup:
		return d.chatStrategy
	default:
		return d.userStrategy
	}
}
