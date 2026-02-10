package channel

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/provider"
	"github.com/gotomicro/ego/core/elog"
)

type baseChannel struct {
	builder provider.SelectorBuilder
	logger  *elog.Component
}

func (s *baseChannel) Send(ctx context.Context, n notification.Notification) (notification.NotificationResponse, error) {
	selector, err := s.builder.Build()
	if err != nil {
		return notification.NotificationResponse{}, fmt.Errorf("%s: %w", "发送通知失败", err)
	}

	for {
		p, nextErr := selector.Next(ctx, n)
		if nextErr != nil {
			return notification.NotificationResponse{}, fmt.Errorf("发送通知失败: %w", nextErr)
		}

		// 使用当前供应商发送
		resp, sendErr := p.Send(ctx, n)
		if sendErr != nil {
			s.logger.Error("使用当前供应商发送失败，将继续向下重试直到结束", elog.FieldErr(sendErr))
			continue
		}

		return resp, nil
	}
}
