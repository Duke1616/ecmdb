package strategy

import (
	"context"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/workflow"
)

type NodeName string

const (
	Start      NodeName = "START"      // 开始节点
	Automation NodeName = "AUTOMATION" // 自动化节点
	User       NodeName = "USER"       // 用户审批节点
)

const (
	// LarkTemplateApprovalName 正常审批通知
	LarkTemplateApprovalName = workflow.NotifyTypeApproval
	// LarkTemplateApprovalRevokeName 带有撤销的审批通知
	LarkTemplateApprovalRevokeName = workflow.NotifyTypeRevoke
	// LarkTemplateCC 抄送通知
	LarkTemplateCC = workflow.NotifyTypeCC
)

type StrategyInfo struct {
	NodeName    NodeName          `json:"node_name"`    // 节点名称
	OrderInfo   order.Order       `json:"order_info"`   // 工单提交信息
	WfInfo      workflow.Workflow `json:"wf_info"`      // 流程信息
	InstanceId  int               `json:"instance_id"`  // 实例 id
	CurrentNode *model.Node       `json:"current_node"` // 流程当前节点
}

// SendStrategy 针对不同节点的策略
type SendStrategy interface {
	Send(ctx context.Context, info StrategyInfo) (notification.NotificationResponse, error)
}

type Dispatcher struct {
	startStrategy      *StartNotification
	automationStrategy *AutomationNotification
	userStrategy       *UserNotification
}

func NewDispatcher(
	startStrategy *StartNotification,
	automationStrategy *AutomationNotification,
	userStrategy *UserNotification,
) SendStrategy {
	return &Dispatcher{
		startStrategy:      startStrategy,
		automationStrategy: automationStrategy,
		userStrategy:       userStrategy,
	}
}

func (d *Dispatcher) Send(ctx context.Context, info StrategyInfo) (notification.NotificationResponse, error) {
	return d.selectStrategy(info).Send(ctx, info)
}

func (d *Dispatcher) selectStrategy(not StrategyInfo) SendStrategy {
	switch not.NodeName {
	case User:
		return d.userStrategy
	case Automation:
		return d.automationStrategy
	case Start:
		return d.startStrategy
	default:
		return d.userStrategy
	}
}
