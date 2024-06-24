package event

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mqx"
	"github.com/ecodeclub/mq-api"
)

type TaskRunnerEventProducer interface {
	Produce(ctx context.Context, topic string, evt RunnerEvent) error
	AddProducer(topic string) error
	DelProducer(topic string) error
}

func NewTaskRunnerEventProducer(q mq.MQ) (TaskRunnerEventProducer, error) {
	return mqx.NewMultipleProducer[RunnerEvent](q)
}
