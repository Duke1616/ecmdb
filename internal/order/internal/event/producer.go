package event

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mqx"
	"github.com/ecodeclub/mq-api"
)

type CreateProcessEventProducer interface {
	Produce(ctx context.Context, evt OrderEvent) error
}

func NewCreateProcessEventProducer(q mq.MQ) (CreateProcessEventProducer, error) {
	return mqx.NewGeneralProducer[OrderEvent](q, CreateProcessEventName)
}
