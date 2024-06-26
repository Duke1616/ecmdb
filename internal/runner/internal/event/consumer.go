package event

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/runner/internal/domain"
	"github.com/Duke1616/ecmdb/internal/runner/internal/service"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/ecodeclub/mq-api"
	"log/slog"
)

type TaskRunnerConsumer struct {
	svc         service.Service
	consumer    mq.Consumer
	workerSvc   worker.Service
	codebookSvc codebook.Service
}

func NewTaskRunnerConsumer(svc service.Service, mq mq.MQ, workerSvc worker.Service, codebookSvc codebook.Service) (*TaskRunnerConsumer, error) {
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

	//  验证代码模版密钥是否正确
	exist, err := c.codebookSvc.ValidationSecret(ctx, evt.TaskIdentifier, evt.TaskSecret)
	if exist != true {
		slog.Error("runner 注册失败", err)
	}

	// 验证节点是否存在
	exist, err = c.workerSvc.ValidationByName(ctx, evt.WorkName)
	if exist != true {
		slog.Error("runner 注册失败", err)
	}

	// 注册服务
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
		TaskIdentifier: req.TaskIdentifier,
		TaskSecret:     req.TaskSecret,
		WorkName:       req.WorkName,
		Name:           req.Name,
		Tags:           req.Tags,
		Desc:           req.Desc,
		Action:         domain.Action(req.Action),
	}
}
