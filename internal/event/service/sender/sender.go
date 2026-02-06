package sender

import (
	"context"
	"sync"

	"github.com/Duke1616/ecmdb/internal/event/domain"
	"github.com/Duke1616/ecmdb/internal/event/service/channel"
	"github.com/gotomicro/ego/core/elog"
)

type NotificationSender interface {
	Send(ctx context.Context, notification domain.Notification) (bool, error)
	BatchSend(ctx context.Context, notifications []domain.Notification) (bool, error)
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

	return true, nil
}
