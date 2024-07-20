package producer

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/engine/internal/event"
	"github.com/Duke1616/ecmdb/pkg/mqx"
	"github.com/ecodeclub/mq-api"
)

type OrderStatusModifyEventProducer interface {
	Produce(ctx context.Context, evt event.OrderStatusModifyEvent) error
}

func NewOrderStatusModifyEventProducer(q mq.MQ) (OrderStatusModifyEventProducer, error) {
	return mqx.NewGeneralProducer[event.OrderStatusModifyEvent](q, event.OrderStatusModifyEventName)
}
