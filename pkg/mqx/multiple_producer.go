package mqx

import (
	"context"
	"fmt"
	"sync"

	"github.com/ecodeclub/mq-api"
)

// MultipleProducer 管理多个 GeneralProducer
type MultipleProducer[T any] struct {
	mu        sync.RWMutex
	producers map[string]*GeneralProducer[T]
	mq        mq.MQ
}

// NewMultipleProducer 创建一个新的 ProducerManager
func NewMultipleProducer[T any](mq mq.MQ) (*MultipleProducer[T], error) {
	return &MultipleProducer[T]{
		producers: make(map[string]*GeneralProducer[T]),
		mq:        mq,
	}, nil
}

// AddProducer 添加一个新的 producer 监听指定的 topic
func (pm *MultipleProducer[T]) AddProducer(topic string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.producers[topic]; exists {
		return fmt.Errorf("topic %s 已经存在", topic)
	}

	producer, err := NewGeneralProducer[T](pm.mq, topic)
	if err != nil {
		return err
	}

	pm.producers[topic] = producer
	return nil
}

// DelProducer 删除 producer 监听指定的 topic
func (pm *MultipleProducer[T]) DelProducer(topic string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.producers[topic]; !exists {
		return fmt.Errorf("topic %s 不存在", topic)
	}

	delete(pm.producers, topic)
	return nil
}

// Produce 发送消息到指定的 topic
func (pm *MultipleProducer[T]) Produce(ctx context.Context, topic string, evt T) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	producer, exists := pm.producers[topic]
	if !exists {
		return fmt.Errorf("topic %s 未找到", topic)
	}

	return producer.Produce(ctx, evt)
}
