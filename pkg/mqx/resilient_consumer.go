package mqx

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
)

// ErrConsumerClosed 表示封装器已经被服务主动关闭。
var ErrConsumerClosed = errors.New("message consumer is closed")

const (
	minReconnectDelay = time.Second
	maxReconnectDelay = 30 * time.Second
)

// ResilientConsumer 在 MQ Consumer 因长时间断网失效后负责退避重建。
type ResilientConsumer struct {
	mq        mq.MQ
	topic     string
	group     string
	lifecycle context.Context
	cancel    context.CancelFunc

	mu      sync.Mutex
	current mq.Consumer
	closed  bool
	logger  *elog.Component
}

func NewResilientConsumer(q mq.MQ, topic, group string) *ResilientConsumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &ResilientConsumer{
		mq: q, topic: topic, group: group, lifecycle: ctx, cancel: cancel,
		logger: elog.DefaultLogger.With(
			elog.FieldComponentName("mq.consumer"),
			elog.String("topic", topic), elog.String("group", group),
		),
	}
}

func (c *ResilientConsumer) Consume(ctx context.Context) (*mq.Message, error) {
	delay := minReconnectDelay
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		consumer, err := c.getOrCreate()
		if errors.Is(err, ErrConsumerClosed) {
			return nil, err
		}
		if err != nil {
			c.logger.Warn("创建 MQ 消费者失败，准备重试", elog.FieldErr(err), elog.Duration("backoff", delay))
		}
		if err == nil {
			attemptCtx, cancel := c.withLifecycleContext(ctx)
			message, consumeErr := consumer.Consume(attemptCtx)
			cancel()
			if consumeErr == nil {
				return message, nil
			}
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if c.isClosed() {
				return nil, ErrConsumerClosed
			}
			c.logger.Warn("MQ 消费者失效，准备重建", elog.FieldErr(consumeErr), elog.Duration("backoff", delay))
			c.invalidate(consumer)
		}
		if err = waitReconnect(ctx, c.lifecycle, delay); err != nil {
			return nil, err
		}
		delay = nextReconnectDelay(delay)
	}
}

// ConsumeChan keeps compatibility with mq.Consumer while retaining reconnect behavior.
func (c *ResilientConsumer) ConsumeChan(ctx context.Context) (<-chan *mq.Message, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if closed {
		return nil, ErrConsumerClosed
	}
	messages := make(chan *mq.Message)
	go func() {
		defer close(messages)
		for {
			message, err := c.Consume(ctx)
			if err != nil {
				return
			}
			select {
			case messages <- message:
			case <-ctx.Done():
				return
			}
		}
	}()
	return messages, nil
}

func (c *ResilientConsumer) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	consumer := c.current
	c.current = nil
	c.cancel()
	c.mu.Unlock()
	if consumer == nil {
		return nil
	}
	return consumer.Close()
}

func (c *ResilientConsumer) getOrCreate() (mq.Consumer, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, ErrConsumerClosed
	}
	if c.current != nil {
		return c.current, nil
	}
	consumer, err := c.mq.Consumer(c.topic, c.group)
	if err != nil {
		return nil, err
	}
	c.current = consumer
	c.logger.Info("MQ 消费者已创建或重建")
	return consumer, nil
}

func (c *ResilientConsumer) invalidate(consumer mq.Consumer) {
	c.mu.Lock()
	if c.current != consumer {
		c.mu.Unlock()
		return
	}
	c.current = nil
	c.mu.Unlock()
	_ = consumer.Close()
}

func (c *ResilientConsumer) isClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

func (c *ResilientConsumer) withLifecycleContext(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	stop := context.AfterFunc(c.lifecycle, cancel)
	return ctx, func() {
		stop()
		cancel()
	}
}

func waitReconnect(ctx, lifecycle context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-lifecycle.Done():
		return ErrConsumerClosed
	case <-timer.C:
		return nil
	}
}

func nextReconnectDelay(delay time.Duration) time.Duration {
	delay *= 2
	if delay > maxReconnectDelay {
		return maxReconnectDelay
	}
	return delay
}
