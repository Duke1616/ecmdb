package strategy

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/event/errs"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/sender"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/gotomicro/ego/core/elog"
)

type AutomationNotification struct {
	sender      sender.NotificationSender
	templateSvc template.Service
	resultSvc   FetcherResult
	userSvc     user.Service
	logger      *elog.Component
}

func NewAutomationNotification(resultSvc FetcherResult, userSvc user.Service, templateSvc template.Service,
	sender sender.NotificationSender) (*AutomationNotification, error) {
	return &AutomationNotification{
		sender:      sender,
		resultSvc:   resultSvc,
		templateSvc: templateSvc,
		userSvc:     userSvc,
		logger:      elog.DefaultLogger,
	}, nil
}

func (n *AutomationNotification) Send(ctx context.Context, info StrategyInfo) (notification.NotificationResponse, error) {
	// 获取当前节点信息
	property, err := getNodeProperty[easyflow.AutomationProperty](info.WfInfo, info.CurrentNode.NodeID)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), err.Error()), fmt.Errorf("%w: %v", errs.ErrNodeNotConfigured, err)
	}

	// 判断是否开启消息发送，以及是否为立即发送
	if !property.IsNotify {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), "【自动化节点】未配置消息通知"), fmt.Errorf("%w", errs.ErrNodeNotConfigured)
	}

	// 判断模式如果不是理解发送则退出
	if !containsAutoNotifyMethod(property.NotifyMethod, ProcessNowSend) {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), "【自动化节点】节点未开启消息通知"), fmt.Errorf("%w", errs.ErrNodeNotConfigured)
	}

	// 查看返回的消息
	wantResult, err := n.resultSvc.FetchResult(ctx, info.InstanceId, info.CurrentNode.NodeID)
	if err != nil {
		n.logger.Warn("执行错误或未开启消息通知",
			elog.FieldErr(err),
			elog.Any("instId", info.InstanceId),
		)
		return notification.NewErrorResponse(string(errs.ErrorCodeFetchDataFailed), err.Error()), fmt.Errorf("%w: %v", errs.ErrFetchData, err)
	}

	startUser, err := n.userSvc.FindByUsername(ctx, info.OrderInfo.CreateBy)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeFetchDataFailed), err.Error()), fmt.Errorf("%w: %v", errs.ErrFetchData, err)
	}

	// 获取模版名称
	tName, err := n.getTemplateName(ctx, info.OrderInfo)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeFetchDataFailed), err.Error()), fmt.Errorf("%w: %v", errs.ErrFetchData, err)
	}

	return n.sender.Send(ctx, notification.Notification{
		Channel:    notification.ChannelLarkCard,
		WorkFlowID: info.WfInfo.Id,
		Receiver:   startUser.FeishuInfo.UserId,
		Template: notification.Template{
			Name:     LarkTemplateApprovalName,
			Title:    rule.GenerateAutoTitle("你提交", tName),
			Fields:   n.getFields(wantResult),
			Values:   []notification.Value{},
			HideForm: true,
		},
	})
}

func (n *AutomationNotification) getTemplateName(ctx context.Context, order order.Order) (string, error) {
	// 获取模版详情信息
	t, err := n.templateSvc.DetailTemplate(ctx, order.TemplateId)
	if err != nil {
		return t.Name, err
	}

	return "", nil
}

func (n *AutomationNotification) getFields(wantResult map[string]interface{}) []notification.Field {
	num := 1
	var fields []notification.Field

	for field, value := range wantResult {
		title := field

		fields = append(fields, notification.Field{
			IsShort: true,
			Tag:     "lark_md",
			Content: fmt.Sprintf(`**%s:**\n%v`, title, value),
		})

		if num%2 == 0 {
			fields = append(fields, notification.Field{
				IsShort: false,
				Tag:     "lark_md",
				Content: "",
			})
		}

		num++
	}

	return fields
}
