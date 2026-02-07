package channel

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/event/domain"
)

type Channel interface {
	// Send 发送通知
	Send(ctx context.Context, notification domain.Notification) (domain.NotificationResponse, error)
}

// Dispatcher 渠道分发器，对外伪装成Channel，作为统一入口
type Dispatcher struct {
	channels map[domain.Channel]Channel
}

// NewDispatcher 创建渠道分发器
func NewDispatcher(channels map[domain.Channel]Channel) *Dispatcher {
	return &Dispatcher{
		channels: channels,
	}
}

func (d *Dispatcher) Send(ctx context.Context, notification domain.Notification) (domain.NotificationResponse, error) {
	channel, ok := d.channels[notification.Channel]
	if !ok {
		return domain.NotificationResponse{}, fmt.Errorf("%s: %s", "无可用通知渠道", notification.Channel)
	}
	return channel.Send(ctx, notification)
}
