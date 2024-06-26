package event

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/runner/internal/domain"
	"github.com/Duke1616/ecmdb/internal/runner/internal/service"
	"github.com/ecodeclub/mq-api"
	"log/slog"
)

type TaskRunnerConsumer struct {
	svc      service.Service
	consumer mq.Consumer
}

func NewTaskRunnerConsumer(svc service.Service, mq mq.MQ) (*TaskRunnerConsumer, error) {
	groupID := "task_runner"
	consumer, err := mq.Consumer(TaskRunnerEventName, groupID)
	if err != nil {
		return nil, err
	}
	return &TaskRunnerConsumer{
		svc:      svc,
		consumer: consumer,
	}, nil
}

func (c *TaskRunnerConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				slog.Error("同步事件失败", err)
			}
		}
	}()
}

func (c *TaskRunnerConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	var evt TaskRunnerEvent
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	if _, err = c.svc.Register(ctx, c.toDomain(evt)); err != nil {
		slog.Error("runner 注册失败", err)
	}

	return err
}

func (c *TaskRunnerConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}

func (c *TaskRunnerConsumer) toDomain(req TaskRunnerEvent) domain.Runner {
	return domain.Runner{
		TaskIdentifier: req.TaskIdentify,
		TaskSecret:     req.TaskSecret,
		WorkName:       req.WorkName,
		Name:           req.Name,
		Tags:           req.Tags,
		Desc:           req.Desc,
		Action:         domain.Action(req.Action),
	}
}
