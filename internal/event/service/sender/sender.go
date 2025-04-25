package sender

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/event/domain"
	"github.com/Duke1616/ecmdb/internal/event/service/channel"
	"github.com/gotomicro/ego/core/elog"
	"sync"
)

type NotificationSender interface {
	Send(ctx context.Context, notification domain.Notification) (bool, error)
	BatchSend(ctx context.Context, notifications []domain.Notification) (bool, error)
}

type sender struct {
	channel channel.Channel
	logger  *elog.Component
}

// NewSender 创建通知发送器
func NewSender(channel channel.Channel) NotificationSender {
	return &sender{
		channel: channel,
		logger:  elog.DefaultLogger,
	}
}

// Send 单条发送通知
func (d *sender) Send(ctx context.Context, notifications domain.Notification) (bool, error) {
	_, err := d.channel.Send(ctx, notifications)
	if err != nil {
		d.logger.Error("发送失败", elog.FieldErr(err))
		return false, err
	}

	return true, nil
}

// BatchSend 批量发送通知
func (d *sender) BatchSend(ctx context.Context, notifications []domain.Notification) (bool, error) {
	if len(notifications) == 0 {
		return false, nil
	}

	var wg sync.WaitGroup
	wg.Add(len(notifications))
	for i := range notifications {
		n := notifications[i]
		go func() {
			defer wg.Done()
			_, err := d.channel.Send(ctx, n)
			if err != nil {
				d.logger.Error("发送失败", elog.FieldErr(err))
			}
		}()
	}
	wg.Wait()

	return true, nil
}
