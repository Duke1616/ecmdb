package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/task/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
)

type ExecuteResultConsumer struct {
	consumer mq.Consumer
	svc      service.Service
	logger   *elog.Component
}

func NewExecuteResultConsumer(q mq.MQ, svc service.Service) (*ExecuteResultConsumer, error) {
	groupID := "task_receive_execute"
	consumer, err := q.Consumer(ExecuteResultEventName, groupID)
	if err != nil {
		return nil, err
	}

	return &ExecuteResultConsumer{
		consumer: consumer,
		svc:      svc,
		logger:   elog.DefaultLogger,
	}, nil
}

func (c *ExecuteResultConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				c.logger.Error("同步修改任务执行状态失败", elog.Any("错误信息", err))
				time.Sleep(time.Second)
			}
		}
	}()
}

func (c *ExecuteResultConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}
	var evt ExecuteResultEvent
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	triggerPosition := domain.TriggerPositionTaskExecutionSuccess
	if domain.Status(evt.Status) == domain.FAILED {
		triggerPosition = domain.TriggerPositionTaskExecutionFailed
	}

	_, err = c.svc.UpdateTaskResult(ctx, domain.TaskResult{
		Id:              evt.TaskId,
		Result:          evt.Result,
		WantResult:      evt.WantResult,
		TriggerPosition: triggerPosition.ToString(),
		Status:          domain.Status(evt.Status),
	})

	return err
}

func (c *ExecuteResultConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}
