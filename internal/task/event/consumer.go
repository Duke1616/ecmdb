package event

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
)

type TaskEventConsumer struct {
	consumer mq.Consumer
	logger   *elog.Component
}

func NewTaskEventConsumer(q mq.MQ) (*TaskEventConsumer, error) {
	groupID := "task"
	consumer, err := q.Consumer(CreateFLowEventName, groupID)
	if err != nil {
		return nil, err
	}
	return &TaskEventConsumer{
		consumer: consumer,
		logger:   elog.DefaultLogger,
	}, nil
}

func (c *TaskEventConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				elog.Error("同步事件失败", elog.Any("err", err))
			}
		}
	}()
}

func (c *TaskEventConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	var evt OrderEvent
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	fmt.Println(evt)
	return nil
}

func (c *TaskEventConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}
