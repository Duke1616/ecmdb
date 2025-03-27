package node

import (
	"context"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/event/easyflow/notification/method"
	"github.com/Duke1616/ecmdb/internal/event/easyflow/notification/result"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/enotify/notify"
	"github.com/gotomicro/ego/core/elog"
)

type AutomationNotification struct {
	integrations []method.NotifyIntegration
	resultSvc    result.FetcherResult
	userSvc      user.Service
	logger       *elog.Component
}

func NewAutomationNotification(resultSvc result.FetcherResult, userSvc user.Service, integrations []method.NotifyIntegration) (*AutomationNotification, error) {
	return &AutomationNotification{
		integrations: integrations,
		resultSvc:    resultSvc,
		userSvc:      userSvc,
		logger:       elog.DefaultLogger,
	}, nil
}

func (n *AutomationNotification) Send(ctx context.Context, nOrder order.Order, wf workflow.Workflow,
	instanceId int, currentNode *model.Node) (bool, error) {
	// 获取当前节点信息
	property, err := getNodeProperty[easyflow.AutomationProperty](wf, currentNode.NodeID)
	if err != nil {
		return false, err
	}

	// 判断是否开启消息发送，以及是否为立即发送
	if !property.IsNotify {
		n.logger.Warn("【自动化节点】未配置消息通知")
		return false, nil
	}

	// 判断模式如果不是理解发送则退出
	if !containsAutoNotifyMethod(property.NotifyMethod, ProcessNowSend) {
		n.logger.Warn("【自动化节点】节点未开启消息通知")
		return false, nil
	}

	// 查看返回的消息
	wantResult, err := n.resultSvc.FetchResult(ctx, instanceId, currentNode.NodeID)
	if err != nil {
		n.logger.Warn("执行错误或未开启消息通知",
			elog.FieldErr(err),
			elog.Any("instId", instanceId),
		)
		return false, err
	}

	// 获取工单创建用户
	startUser, err := n.userSvc.FindByUsername(ctx, nOrder.CreateBy)
	if err != nil {
		return false, err
	}

	var messages []notify.NotifierWrap
	title := rule.GenerateAutoTitle("你提交", nOrder.TemplateName)
	for _, integration := range n.integrations {
		if integration.Name == fmt.Sprintf("%s_%s", workflow.NotifyMethodToString(wf.NotifyMethod), "automation") {
			messages = integration.Notifier.Builder(title, []user.User{startUser},
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
			elog.Any("user", startUser.DisplayName),
		)
	}

	return true, nil
}
