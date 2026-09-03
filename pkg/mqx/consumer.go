package mqx

import (
	"context"
	"sync"

	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
)

type (
	ConsumeFunc func(ctx context.Context, message *mq.Message) error
	Consumer    struct {
		name     string
		consumer *ResilientConsumer
		ctx      context.Context
		cancel   context.CancelFunc
		wg       sync.WaitGroup

		logger *elog.Component
	}
)

func NewConsumer(name string, mq mq.MQ, topic string) *Consumer {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &Consumer{
		name:     name,
		consumer: NewResilientConsumer(mq, topic, name),
		ctx:      ctx,
		cancel:   cancelFunc,
		logger:   elog.DefaultLogger.With(elog.FieldComponent(name)),
	}
}

func (c *Consumer) Start(ctx context.Context, consumeFunc ConsumeFunc) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.consume(ctx, consumeFunc)
	}()
	return nil
}

func (c *Consumer) consume(ctx context.Context, consumeFunc func(ctx context.Context, message *mq.Message) error) {
	c.logger.Info("消费者已启动")
	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("内部上下文取消，结束消费循环")
			return
		case <-ctx.Done():
			c.logger.Info("参数上下文取消，结束消费循环")
			return
		default:
			message, err := c.consumer.Consume(ctx)
			if err != nil {
				return
			}
			activeCtx := ExtractContext(ctx, message)
			err = consumeFunc(activeCtx, message)
			if err != nil {
				c.logger.Error("消费消息失败",
					elog.String("消息体", string(message.Value)),
					elog.FieldErr(err))
			}
			c.logger.Info("消费消息成功",
				elog.String("消息体", string(message.Value)),
			)
		}
	}
}

func (c *Consumer) Name() string {
	return c.name
}

func (c *Consumer) Stop() error {
	c.cancel()
	err := c.consumer.Close()
	c.wg.Wait()
	return err
}
