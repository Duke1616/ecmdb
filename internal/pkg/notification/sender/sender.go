package sender

import (
	"context"
	"sync"

	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/channel"
	"github.com/gotomicro/ego/core/elog"
)

//go:generate mockgen -source=./sender.go -package=sendermocks -destination=../../mocks/sender.mock.go -typed NotificationSender
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
		d.logger.Error("发送失败", elog.FieldErr(err), elog.String("receiver", n.Receiver))
		return notification.NotificationResponse{}, err
	}

	d.logger.Info("【Notification】通知发送成功",
		elog.String("receiver", n.Receiver),
		elog.String("receiver_type", n.ReceiverType),
		elog.String("template", string(n.Template.Name)),
		elog.String("channel", n.Channel.String()))

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
				d.logger.Error("发送失败", elog.FieldErr(err), elog.String("receiver", n.Receiver))
			} else {
				d.logger.Info("【Notification】通知发送成功",
					elog.String("receiver", n.Receiver),
					elog.String("receiver_type", n.ReceiverType),
					elog.String("template", string(n.Template.Name)),
					elog.String("channel", n.Channel.String()))
			}
		}()
	}
	wg.Wait()
	d.logger.Info("【Notification】批量通知发送处理完成", elog.Int("total", len(notifications)))

	return notification.NotificationResponse{}, nil
}
