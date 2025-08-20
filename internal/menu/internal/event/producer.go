package event

import (
	"context"

	"github.com/Duke1616/ecmdb/pkg/mqx"
	"github.com/ecodeclub/mq-api"
)

type MenuChangeEventProducer interface {
	Produce(ctx context.Context, evt MenuEvent) error
}

func NewMenuChangeEventProducer(q mq.MQ) (MenuChangeEventProducer, error) {
	return mqx.NewGeneralProducer[MenuEvent](q, MenuChangeEventName)
}
