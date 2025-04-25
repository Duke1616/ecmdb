package channel

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/event/domain"
	"github.com/Duke1616/ecmdb/internal/event/service/provider"
)

type baseChannel struct {
	builder provider.SelectorBuilder
}

func (s *baseChannel) Send(ctx context.Context, notification domain.Notification) (bool, error) {
	selector, err := s.builder.Build()
	if err != nil {
		return false, fmt.Errorf("%s: %w", "发送通知失败", err)
	}

	for {
		p, err1 := selector.Next(ctx, notification)
		if err1 != nil {
			return false, fmt.Errorf("%s: %w", "发送通知失败", err1)
		}
		
		return p.Send(ctx, notification)
	}
}
