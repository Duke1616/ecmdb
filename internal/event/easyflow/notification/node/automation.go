package node

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/event/easyflow/notification"
	"github.com/Duke1616/ecmdb/internal/event/easyflow/notification/method"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/enotify/notify"
	"github.com/gotomicro/ego/core/elog"
)

type AutomationNotification[T any] struct {
	integrations []method.NotifyIntegration
	taskSvc      task.Service
	userSvc      user.Service
	logger       *elog.Component
}

func NewAutomationNotification[T any](taskSvc task.Service, userSvc user.Service, integrations []method.NotifyIntegration) (*AutomationNotification[T], error) {
	return &AutomationNotification[T]{
		integrations: integrations,
		taskSvc:      taskSvc,
		userSvc:      userSvc,
		logger:       elog.DefaultLogger,
	}, nil
}

func (n *AutomationNotification[T]) UnmarshalProperty(ctx context.Context, wf workflow.Workflow, nodeId string) (T, error) {
	//TODO implement me
	panic("implement me")
}

func (n *AutomationNotification[T]) Send(ctx context.Context, nOrder order.Order,
	params notification.NotifyParams) (bool, error) {

	u, err := n.userSvc.FindByUsername(ctx, nOrder.CreateBy)
	if err != nil {
		return false, err
	}

	var messages []notify.NotifierWrap
	for _, integration := range n.integrations {
		if integration.Name == fmt.Sprintf("%s_%s", params.NotifyMethod, "automation") {
			messages = integration.Notifier.Builder("自动化任务返回结果", []user.User{u}, method.NotifyParams{
				Order:      nOrder,
				WantResult: params.WantResult,
			})
			break
		}
	}

	var ok bool
	if ok, err = send(context.Background(), messages); err != nil || !ok {
		n.logger.Warn("发送消息失败",
			elog.Any("error", err),
			elog.Any("user", u.DisplayName),
		)
	}

	return true, nil
}

func (n *AutomationNotification[T]) IsNotification(ctx context.Context, wf workflow.Workflow, instanceId int,
	nodeId string) (bool, map[string]interface{}, error) {
	nodesJSON, err := json.Marshal(wf.FlowData.Nodes)
	if err != nil {
		return false, nil, err
	}
	var nodes []easyflow.Node
	err = json.Unmarshal(nodesJSON, &nodes)
	if err != nil {
		return false, nil, err
	}

	property := easyflow.AutomationProperty{}
	for _, node := range nodes {
		if node.ID == nodeId {
			property, _ = easyflow.ToNodeProperty[easyflow.AutomationProperty](node)
		}
	}

	// 判断是否开启消息发送，以及是否为立即发送
	if !property.IsNotify || property.NotifyMethod != ProcessNowSend {
		return false, nil, nil
	}

	result, err := n.taskSvc.FindTaskResult(ctx, instanceId, nodeId)
	if err != nil {
		return false, nil, err
	}

	if result.WantResult == "" {
		return false, nil, fmt.Errorf("返回值为空, 不做任何数据处理")
	}

	var wantResult map[string]interface{}
	err = json.Unmarshal([]byte(result.WantResult), &wantResult)
	if err != nil {
		return false, nil, err
	}

	return true, wantResult, nil
}
