package event

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mqx"
	"github.com/ecodeclub/mq-api"
)

type CreateFlowEventProducer interface {
	Produce(ctx context.Context, evt OrderEvent) error
}

func NewCreateFlowEventProducer(q mq.MQ) (CreateFlowEventProducer, error) {
	return mqx.NewGeneralProducer[OrderEvent](q, CreateFLowEventName)
}
