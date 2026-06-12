package mqx

import (
	"context"

	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
)

type (
	ConsumeFunc func(ctx context.Context, message *mq.Message) error
	Consumer    struct {
		name string

		mq     mq.MQ
		topic  string
		ctx    context.Context
		cancel context.CancelFunc

		logger *elog.Component
	}
)

func NewConsumer(name string, mq mq.MQ, topic string) *Consumer {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &Consumer{
		name:   name,
		mq:     mq,
		topic:  topic,
		ctx:    ctx,
		cancel: cancelFunc,
		logger: elog.DefaultLogger.With(elog.FieldComponent(name)),
	}
}

func (c *Consumer) Start(ctx context.Context, consumeFunc ConsumeFunc) error {
	consumer, err := c.mq.Consumer(c.topic, c.name)
	if err != nil {
		c.logger.Error("获取MQ消费者失败",
			elog.FieldErr(err),
		)
		return err
	}
	ch, err := consumer.ConsumeChan(ctx)
	if err != nil {
		c.logger.Error("获取MQ消费者Chan失败",
			elog.FieldErr(err),
		)
		return err
	}
	go c.consume(ctx, ch, consumeFunc)
	return nil
}

func (c *Consumer) consume(ctx context.Context, mqChan <-chan *mq.Message, consumeFunc func(ctx context.Context, message *mq.Message) error) {
	c.logger.Info("消费者已启动")
	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("内部上下文取消，结束消费循环")
			return
		case <-ctx.Done():
			c.logger.Info("参数上下文取消，结束消费循环")
			return
		case message, ok := <-mqChan:
			if !ok {
				return
			}
			// 自动从 Kafka 消息头部提取并解包多租户及业务 Context
			activeCtx := ExtractContext(ctx, message)

			err := consumeFunc(activeCtx, message)
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
	return nil
}
