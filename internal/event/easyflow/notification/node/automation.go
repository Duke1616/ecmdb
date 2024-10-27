package node

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/workflow"
)

type AutomationNotification struct {
	orderSvc    order.Service
	workflowSvc workflow.Service
}

func (n *AutomationNotification) Send(ctx context.Context, instanceId int, userIDs []string) (bool, error) {
	return true, nil
}

func (n *AutomationNotification) IsNotification(ctx context.Context, InstanceId int) (order.Order, workflow.NotifyMethod, bool, error) {
	// 获取工单详情信息
	o, err := n.orderSvc.DetailByProcessInstId(ctx, InstanceId)
	if err != nil {
		return o, 0, false, err
	}

	// 判断是否需要消息提示
	wf, err := n.workflowSvc.Find(ctx, o.WorkflowId)
	if err != nil {
		return o, 0, false, err
	}

	if !wf.IsNotify {
		return order.Order{}, 0, false, nil
	}

	return o, wf.NotifyMethod, true, nil
}
