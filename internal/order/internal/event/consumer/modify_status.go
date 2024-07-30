package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/order/internal/event"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
)

type OrderStatusModifyEventConsumer struct {
	svc      service.Service
	consumer mq.Consumer
	logger   *elog.Component
}

func NewOrderStatusModifyEventConsumer(q mq.MQ, svc service.Service) (*OrderStatusModifyEventConsumer, error) {
	groupID := "order_status_modify"
	consumer, err := q.Consumer(event.OrderStatusModifyEventName, groupID)
	if err != nil {
		return nil, err
	}

	return &OrderStatusModifyEventConsumer{
		consumer: consumer,
		svc:      svc,
		logger:   elog.DefaultLogger,
	}, nil
}

func (c *OrderStatusModifyEventConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				c.logger.Error("同步事件失败", elog.Any("err", err))
			}
		}
	}()
}

func (c *OrderStatusModifyEventConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}
	var evt event.OrderStatusModifyEvent
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	return c.svc.UpdateStatusByInstanceId(ctx, evt.ProcessInstanceId, evt.Status.ToUint8())
}

func (c *OrderStatusModifyEventConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}
