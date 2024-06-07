package mqx

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ecodeclub/mq-api"
)

type Producer[T any] interface {
	Produce(ctx context.Context, evt T) error
}

type GeneralProducer[T any] struct {
	producer mq.Producer
	topic    string
}

func NewGeneralProducer[T any](q mq.MQ, topic string) (*GeneralProducer[T], error) {
	p, err := q.Producer(topic)
	return &GeneralProducer[T]{
		producer: p,
		topic:    topic,
	}, err
}

func (p *GeneralProducer[T]) Produce(ctx context.Context, evt T) error {
	data, err := json.Marshal(&evt)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}
	_, err = p.producer.Produce(ctx, &mq.Message{Value: data})
	if err != nil {
		return fmt.Errorf("向topic=%s发送event=%#v失败: %w", p.topic, evt, err)
	}
	return nil
}
