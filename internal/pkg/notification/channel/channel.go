package channel

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/pkg/notification"
)

type Channel interface {
	// Send 发送通知
	Send(ctx context.Context, n notification.Notification) (notification.NotificationResponse, error)
}

// Dispatcher 渠道分发器，对外伪装成Channel，作为统一入口
type Dispatcher struct {
	channels map[notification.Channel]Channel
}

// NewDispatcher 创建渠道分发器
func NewDispatcher(channels map[notification.Channel]Channel) *Dispatcher {
	return &Dispatcher{
		channels: channels,
	}
}

func (d *Dispatcher) Send(ctx context.Context, n notification.Notification) (notification.NotificationResponse, error) {
	channel, ok := d.channels[n.Channel]
	if !ok {
		return notification.NotificationResponse{}, fmt.Errorf("%s: %s", "无可用通知渠道", n.Channel)
	}
	return channel.Send(ctx, n)
}
