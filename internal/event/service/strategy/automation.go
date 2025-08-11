package strategy

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/event/domain"
	"github.com/Duke1616/ecmdb/internal/event/service/sender"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/enotify/notify/feishu/card"
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

func (n *AutomationNotification) Send(ctx context.Context, notification domain.StrategyInfo) (bool, error) {
	// 获取当前节点信息
	property, err := getNodeProperty[easyflow.AutomationProperty](notification.WfInfo, notification.CurrentNode.NodeID)
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
	wantResult, err := n.resultSvc.FetchResult(ctx, notification.InstanceId, notification.CurrentNode.NodeID)
	if err != nil {
		n.logger.Warn("执行错误或未开启消息通知",
			elog.FieldErr(err),
			elog.Any("instId", notification.InstanceId),
		)
		return false, err
	}

	// 获取工单创建用户
	startUser, err := n.userSvc.FindByUsername(ctx, notification.OrderInfo.CreateBy)
	if err != nil {
		return false, err
	}

	// 获取模版名称
	tName, err := n.getTemplateName(ctx, notification.OrderInfo)
	if err != nil {
		return false, err
	}

	return n.sender.Send(ctx, domain.Notification{
		Channel:  domain.ChannelFeishuCard,
		Receiver: startUser.FeishuInfo.UserId,
		Template: domain.Template{
			Name:     FeishuTemplateApprovalName,
			Title:    rule.GenerateAutoTitle("你提交", tName),
			Fields:   n.getFields(wantResult),
			Values:   []card.Value{},
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

func (n *AutomationNotification) getFields(wantResult map[string]interface{}) []card.Field {
	num := 1
	var fields []card.Field

	for field, value := range wantResult {
		title := field

		fields = append(fields, card.Field{
			IsShort: true,
			Tag:     "lark_md",
			Content: fmt.Sprintf(`**%s:**\n%v`, title, value),
		})

		if num%2 == 0 {
			fields = append(fields, card.Field{
				IsShort: false,
				Tag:     "lark_md",
				Content: "",
			})
		}

		num++
	}

	return fields
}
