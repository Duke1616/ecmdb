package strategy

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/event/errs"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/sender"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/enotify/notify/feishu"
)

type AutomationNotification struct {
	BaseStrategy
	sender sender.NotificationSender
}

func NewAutomationNotification(base BaseStrategy, sender sender.NotificationSender) *AutomationNotification {
	return &AutomationNotification{
		BaseStrategy: base,
		sender:       sender,
	}
}

func (n *AutomationNotification) Send(ctx context.Context, info Info) (notification.NotificationResponse, error) {
	// 1. 获取当前节点信息
	_, rawProps, err := n.GetNodeProperty(info, info.CurrentNode.NodeID)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), err.Error()), fmt.Errorf("%w: %v", errs.ErrNodeNotConfigured, err)
	}

	property, err := easyflow.ToNodeProperty[easyflow.AutomationProperty](easyflow.Node{Properties: rawProps})
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), err.Error()), err
	}

	// 2. 权限与触发校验
	if !n.IsGlobalNotify(info.Workflow) {
		return notification.NewSuccessResponse(0, "全局通知已关闭"), nil
	}

	if !property.IsNotify {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), "【自动化节点】未开启消息通知"), fmt.Errorf("%w", errs.ErrNodeNotConfigured)
	}

	if !containsAutoNotifyMethod(property.NotifyMethod, ProcessNowSend) {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), "【自动化节点】节点未开启消息通知模式"), fmt.Errorf("%w", errs.ErrNodeNotConfigured)
	}

	// 3. 获取元数据与自动化任务结果
	nodes, _, _ := n.GetNodeProperty(info, info.CurrentNode.NodeID)
	data, err := n.FetchRequiredData(ctx, info, nodes)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeFetchDataFailed), err.Error()), fmt.Errorf("%w: %v", errs.ErrFetchData, err)
	}

	// 4. 发送消息
	fields := n.BuildWantResultFields(data.WantResult)

	// 每两个字段插入一个空行（保持原有格式）
	formattedFields := make([]notification.Field, 0, len(fields)*2)
	for i, field := range fields {
		formattedFields = append(formattedFields, field)
		if (i+1)%2 == 0 && i < len(fields)-1 {
			formattedFields = append(formattedFields, notification.Field{
				IsShort: false,
				Tag:     "lark_md",
				Content: "",
			})
		}
	}

	return n.sender.Send(ctx, notification.Notification{
		Channel:      info.Channel,
		ReceiverType: feishu.ReceiveIDTypeUserID,
		WorkFlowID:   info.Workflow.Id,
		Receiver:     data.StartUser.FeishuInfo.UserId,
		Template: notification.Template{
			Name:     LarkTemplateApprovalName,
			Title:    rule.GenerateAutoTitle("你提交", data.TName),
			Fields:   formattedFields,
			HideForm: true,
		},
	})
}
