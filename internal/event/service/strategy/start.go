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
	"github.com/ecodeclub/ekit/slice"
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

func (s *StartNotification) Send(ctx context.Context, info StrategyInfo) (notification.NotificationResponse, error) {
	// 获取当前节点信息
	property, err := getNodeProperty[easyflow.StartProperty](info.WfInfo, info.CurrentNode.NodeID)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), "【开始节点】未配置消息通知"), fmt.Errorf("%w: %v", errs.ErrNodeNotConfigured, err)
	}

	// 判断开始节点是否需要发送消息通知
	if ok := s.isNotify(property, info.InstanceId); !ok {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), "【开始节点】未配置消息通知"), fmt.Errorf("%w", errs.ErrNodeNotConfigured)
	}

	// 解析配置
	rules, tName, err := s.getRules(ctx, info.OrderInfo)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeFetchDataFailed), err.Error()), fmt.Errorf("%w: %v", errs.ErrFetchData, err)
	}

	// 获取工单创建用户
	startUser, err := s.userSvc.FindByUsername(ctx, info.OrderInfo.CreateBy)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeFetchDataFailed), err.Error()), fmt.Errorf("%w: %v", errs.ErrFetchData, err)
	}

	ruleFields := rule.GetFields(rules, info.OrderInfo.Provide.ToUint8(), info.OrderInfo.Data)
	return s.sender.Send(ctx, notification.Notification{
		Channel:    notification.ChannelLarkCard,
		WorkFlowID: info.WfInfo.Id,
		Receiver:   startUser.FeishuInfo.UserId,
		Template: notification.Template{
			Name:  LarkTemplateApprovalRevokeName,
			Title: rule.GenerateTitle("你提交的", tName),
			Fields: slice.Map(ruleFields, func(idx int, src rule.Field) notification.Field {
				return notification.Field{
					IsShort: src.IsShort,
					Tag:     src.Tag,
					Content: src.Content,
				}
			}),
			Values: []notification.Value{
				{
					Key:   "order_id",
					Value: info.OrderInfo.Id,
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
func (s *StartNotification) getRules(ctx context.Context, order order.Order) ([]rule.Rule, string, error) {
	// 获取模版详情信息
	t, err := s.templateSvc.DetailTemplate(ctx, order.TemplateId)
	if err != nil {
		return nil, "", err
	}

	rules, err := rule.ParseRules(t.Rules)

	if err != nil {
		return nil, "", err
	}

	return rules, t.Name, nil
}
