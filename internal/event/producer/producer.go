package producer

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mqx"
	"github.com/ecodeclub/mq-api"
)

type OrderStatusModifyEventProducer interface {
	Produce(ctx context.Context, evt OrderStatusModifyEvent) error
}

func NewOrderStatusModifyEventProducer(q mq.MQ) (OrderStatusModifyEventProducer, error) {
	return mqx.NewGeneralProducer[OrderStatusModifyEvent](q, OrderStatusModifyEventName)
}
