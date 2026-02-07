package channel

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/event/domain"
	"github.com/Duke1616/ecmdb/internal/event/service/provider"
	"github.com/gotomicro/ego/core/elog"
)

type baseChannel struct {
	builder provider.SelectorBuilder
	logger  *elog.Component
}

func (s *baseChannel) Send(ctx context.Context, notification domain.Notification) (domain.NotificationResponse, error) {
	selector, err := s.builder.Build()
	if err != nil {
		return domain.NotificationResponse{}, fmt.Errorf("%s: %w", "发送通知失败", err)
	}

	for {
		p, nextErr := selector.Next(ctx, notification)
		if nextErr != nil {
			return domain.NotificationResponse{}, fmt.Errorf("发送通知失败: %w", nextErr)
		}

		// 使用当前供应商发送
		resp, sendErr := p.Send(ctx, notification)
		if sendErr != nil {
			s.logger.Error("使用当前供应商发送失败，将继续向下重试直到结束", elog.FieldErr(sendErr))
			continue
		}

		return resp, nil
	}
}
