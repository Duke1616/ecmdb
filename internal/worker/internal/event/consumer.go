package event

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/worker/internal/domain"
	"github.com/Duke1616/ecmdb/internal/worker/internal/service"
	"github.com/ecodeclub/mq-api"
	"log/slog"
)

type TaskWorkerConsumer struct {
	svc      service.Service
	consumer mq.Consumer
}

func NewTaskWorkerConsumer(svc service.Service, mq mq.MQ) (*TaskWorkerConsumer, error) {
	groupID := "task_create_worker"
	consumer, err := mq.Consumer(TaskWorkerEventName, groupID)
	if err != nil {
		return nil, err
	}
	return &TaskWorkerConsumer{
		svc:      svc,
		consumer: consumer,
	}, nil
}

func (c *TaskWorkerConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				slog.Error("同步事件失败", err)
			}
		}
	}()
}

func (c *TaskWorkerConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	var evt WorkerEvent
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	if _, err = c.svc.FindOrRegisterByName(ctx, c.toDomain(evt)); err != nil {
		slog.Error("工作节点已经存在或新增工作节点失败", err)
	}

	return err
}

func (c *TaskWorkerConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}

func (c *TaskWorkerConsumer) toDomain(req WorkerEvent) domain.Worker {
	return domain.Worker{
		Name:  req.Name,
		Desc:  req.Desc,
		Topic: req.Topic,
	}
}
