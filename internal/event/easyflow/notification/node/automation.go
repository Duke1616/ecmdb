package node

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/event/easyflow/notification/method"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/enotify/notify"
	"github.com/gotomicro/ego/core/elog"
)

type AutomationNotification struct {
	integrations []method.NotifyIntegration
	taskSvc      task.Service
	userSvc      user.Service
	logger       *elog.Component
}

func NewAutomationNotification(taskSvc task.Service, userSvc user.Service, integrations []method.NotifyIntegration) (*AutomationNotification, error) {
	return &AutomationNotification{
		integrations: integrations,
		taskSvc:      taskSvc,
		userSvc:      userSvc,
		logger:       elog.DefaultLogger,
	}, nil
}

func (n *AutomationNotification) Send(ctx context.Context, nOrder order.Order, wf workflow.Workflow,
	instanceId int, currentNode *model.Node) (bool, error) {
	wantResult, err := n.wantResult(ctx, wf, instanceId, currentNode.NodeID)
	if err != nil {
		n.logger.Warn("执行错误或未开启消息通知",
			elog.FieldErr(err),
			elog.Any("instId", instanceId),
		)
		return false, err
	}

	u, err := n.userSvc.FindByUsername(ctx, nOrder.CreateBy)
	if err != nil {
		return false, err
	}

	var messages []notify.NotifierWrap
	title := rule.GenerateAutoTitle("你提交", nOrder.TemplateName)
	for _, integration := range n.integrations {
		if integration.Name == fmt.Sprintf("%s_%s", workflow.NotifyMethodToString(wf.NotifyMethod), "automation") {
			messages = integration.Notifier.Builder(title, []user.User{u},
				method.FeishuTemplateApprovalName, method.NewNotifyParamsBuilder().
					SetOrder(nOrder).
					SetWantResult(wantResult).
					Build())
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

func (n *AutomationNotification) wantResult(ctx context.Context, wf workflow.Workflow, instanceId int,
	nodeId string) (map[string]interface{}, error) {
	nodesJSON, err := json.Marshal(wf.FlowData.Nodes)
	if err != nil {
		return nil, err
	}
	var nodes []easyflow.Node
	err = json.Unmarshal(nodesJSON, &nodes)
	if err != nil {
		return nil, err
	}

	property := easyflow.AutomationProperty{}
	for _, node := range nodes {
		if node.ID == nodeId {
			property, _ = easyflow.ToNodeProperty[easyflow.AutomationProperty](node)
		}
	}

	// 判断是否开启消息发送，以及是否为立即发送
	if !property.IsNotify || containsNotifyMethod(property.NotifyMethod, ProcessNowSend) {
		return nil, fmt.Errorf("未配置消息通知")
	}

	result, err := n.taskSvc.FindTaskResult(ctx, instanceId, nodeId)
	if err != nil {
		return nil, err
	}

	if result.WantResult == "" {
		return nil, fmt.Errorf("返回值为空, 不做任何数据处理")
	}

	var wantResult map[string]interface{}
	err = json.Unmarshal([]byte(result.WantResult), &wantResult)
	if err != nil {
		return nil, err
	}

	return wantResult, nil
}

func containsNotifyMethod(notifyMethod interface{}, target int64) bool {
	switch v := notifyMethod.(type) {
	case []int64:
		// 如果是数组，遍历检查
		for _, item := range v {
			if item == target {
				return true
			}
		}
		return false
	case int64:
		// 如果是单个值，直接比较
		return v == target
	default:
		// 其他类型（如 nil），返回 false
		return false
	}
}
