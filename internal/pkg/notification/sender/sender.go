package sender

import (
	"context"
	"sync"

	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/channel"
	"github.com/gotomicro/ego/core/elog"
)

type NotificationSender interface {
	Send(ctx context.Context, n notification.Notification) (notification.NotificationResponse, error)
	BatchSend(ctx context.Context, notifications []notification.Notification) (notification.NotificationResponse, error)
}

type sender struct {
	channel channel.Channel
	logger  *elog.Component
	sem     chan struct{}
}

// NewSender 创建通知发送器
func NewSender(channel channel.Channel) NotificationSender {
	return &sender{
		channel: channel,
		logger:  elog.DefaultLogger,
		sem:     make(chan struct{}, 50),
	}
}

// Send 单条发送通知
func (d *sender) Send(ctx context.Context, n notification.Notification) (notification.NotificationResponse, error) {
	_, err := d.channel.Send(ctx, n)
	if err != nil {
		d.logger.Error("发送失败", elog.FieldErr(err))
		return notification.NotificationResponse{}, err
	}

	return notification.NotificationResponse{}, nil
}

// BatchSend 批量发送通知
func (d *sender) BatchSend(ctx context.Context, notifications []notification.Notification) (notification.NotificationResponse, error) {
	if len(notifications) == 0 {
		return notification.NotificationResponse{}, nil
	}

	var wg sync.WaitGroup
	for i := range notifications {
		n := notifications[i]
		wg.Add(1)
		d.sem <- struct{}{} // Acquire semaphore
		go func() {
			defer func() {
				<-d.sem // Release semaphore
				wg.Done()
			}()
			_, err := d.channel.Send(ctx, n)
			if err != nil {
				d.logger.Error("发送失败", elog.FieldErr(err))
			}
		}()
	}
	wg.Wait()

	return notification.NotificationResponse{}, nil
}
