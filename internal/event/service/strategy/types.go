package strategy

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/event/domain"
)

const (
	// LarkTemplateApprovalName 正常审批通知
	LarkTemplateApprovalName = "feishu-card-callback"
	// LarkTemplateApprovalRevokeName 带有撤销的审批通知
	LarkTemplateApprovalRevokeName = "feishu-card-revoke"
	// LarkTemplateCC 抄送通知
	LarkTemplateCC = "feishu-card-cc"
)

// SendStrategy 针对不同节点的策略
type SendStrategy interface {
	Send(ctx context.Context, notification domain.StrategyInfo) (bool, error)
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

func (d *Dispatcher) Send(ctx context.Context, notification domain.StrategyInfo) (bool, error) {
	return d.selectStrategy(notification).Send(ctx, notification)
}

func (d *Dispatcher) selectStrategy(not domain.StrategyInfo) SendStrategy {
	switch not.NodeName {
	case domain.User:
		return d.userStrategy
	case domain.Automation:
		return d.automationStrategy
	case domain.Start:
		return d.startStrategy
	default:
		return d.userStrategy
	}
}
