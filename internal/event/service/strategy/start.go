package strategy

import (
	"context"
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

type StartNotification struct {
	sender      sender.NotificationSender
	userSvc     user.Service
	templateSvc template.Service

	logger *elog.Component
}

func NewStartNotification(userSvc user.Service, templateSvc template.Service, sender sender.NotificationSender) (*StartNotification, error) {
	return &StartNotification{
		sender:      sender,
		userSvc:     userSvc,
		templateSvc: templateSvc,
		logger:      elog.DefaultLogger,
	}, nil
}

func (s *StartNotification) Send(ctx context.Context, notification domain.StrategyInfo) (bool, error) {
	// 获取当前节点信息
	property, err := getNodeProperty[easyflow.StartProperty](notification.WfInfo, notification.CurrentNode.NodeID)
	if err != nil {
		return false, err
	}

	// 判断开始节点是否需要发送消息通知
	if ok := s.isNotify(property, notification.InstanceId); !ok {
		return false, nil
	}

	// 解析配置
	rules, err := s.getRules(ctx, notification.OrderInfo)
	if err != nil {
		return false, err
	}

	// 获取工单创建用户
	startUser, err := s.userSvc.FindByUsername(ctx, notification.OrderInfo.CreateBy)
	if err != nil {
		return false, err
	}

	return s.sender.Send(ctx, domain.Notification{
		Channel:  domain.ChannelFeishuCard,
		Receiver: startUser.FeishuInfo.UserId,
		Template: domain.Template{
			Name:   FeishuTemplateApprovalRevokeName,
			Title:  rule.GenerateTitle("你提交的", notification.OrderInfo.TemplateName),
			Fields: rule.GetFields(rules, notification.OrderInfo.Provide.ToUint8(), notification.OrderInfo.Data),
			Values: []card.Value{
				{
					Key:   "order_id",
					Value: notification.OrderInfo.Id,
				},
				{
					Key:   "task_id",
					Value: "100001",
				},
			},
			HideForm: false,
		},
	})
}

func (s *StartNotification) isNotify(sp easyflow.StartProperty, instanceId int) bool {
	if !sp.IsNotify {
		s.logger.Warn("流程控制【开始节点】未开启消息通知能力",
			elog.Any("instId", instanceId),
		)
		return false
	}

	return true
}

// isNotify 获取模版的字段信息
func (s *StartNotification) getRules(ctx context.Context, order order.Order) ([]rule.Rule, error) {
	// 获取模版详情信息
	t, err := s.templateSvc.DetailTemplate(ctx, order.TemplateId)
	if err != nil {
		return nil, err
	}

	rules, err := rule.ParseRules(t.Rules)

	if err != nil {
		return nil, err
	}

	return rules, nil
}
