package strategy

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/event/errs"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/sender"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	"github.com/gotomicro/ego/core/elog"
)

type StartNotification struct {
	BaseStrategy
	sender sender.NotificationSender
}

func NewStartNotification(base BaseStrategy, sender sender.NotificationSender) *StartNotification {
	return &StartNotification{
		BaseStrategy: base,
		sender:       sender,
	}
}

func (s *StartNotification) Send(ctx context.Context, info Info) (notification.NotificationResponse, error) {
	// 1. 全局通知校验
	if !s.IsGlobalNotify(info.Workflow) {
		return notification.NewSuccessResponse(0, "全局通知已关闭"), nil
	}

	// 2. 开始节点通常通知发起人
	s.Logger.Debug("开始节点发送通知",
		elog.Int("instance_id", info.InstID),
		elog.String("node_id", info.CurrentNode.NodeID))

	// 3. 加载基础通知元数据
	nodes, _, err := s.GetNodeProperty(info, info.CurrentNode.NodeID)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), err.Error()), fmt.Errorf("%w: %v", errs.ErrNodeNotConfigured, err)
	}

	data, err := s.FetchRequiredData(ctx, info, nodes)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeFetchDataFailed), err.Error()), err
	}

	title := rule.GenerateTitle(data.StartUser.DisplayName, data.TName)
	fields := rule.GetFields(data.Rules, info.Order.Provide.ToUint8(), info.Order.Data)

	msg := notification.Notification{
		Channel:    info.Channel,
		WorkFlowID: info.Workflow.Id,
		Receiver:   data.StartUser.FeishuInfo.UserId,
		Template: notification.Template{
			Name:   LarkTemplateCC,
			Title:  title,
			Fields: s.ConvertRuleFields(fields),
			Values: []notification.Value{
				{Key: "order_id", Value: info.Order.Id},
			},
			HideForm: true,
		},
	}

	if msg.Receiver != "" {
		if _, sendErr := s.sender.Send(ctx, msg); sendErr != nil {
			return notification.NotificationResponse{}, sendErr
		}
	}

	return notification.NewSuccessResponse(0, "success"), nil
}
