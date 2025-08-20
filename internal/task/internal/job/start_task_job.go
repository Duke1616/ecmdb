package job

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/ecodeclub/ekit/retry"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/task/ecron"
)

var _ ecron.NamedJob = (*StartTaskJob)(nil)

type StartTaskJob struct {
	svc             service.Service
	initialInterval time.Duration
	maxInterval     time.Duration
	logger          *elog.Component
	maxRetries      int32
	limit           int64
}

func NewStartTaskJob(svc service.Service, limit int64, initialInterval time.Duration,
	maxInterval time.Duration, maxRetries int32) *StartTaskJob {
	return &StartTaskJob{
		svc:             svc,
		logger:          elog.DefaultLogger,
		limit:           limit,
		initialInterval: initialInterval,
		maxInterval:     maxInterval,
		maxRetries:      maxRetries,
	}
}

func (c *StartTaskJob) Name() string {
	return "StartTaskJob"
}

func (c *StartTaskJob) Run(ctx context.Context) error {
	for {
		tasks, total, err := c.svc.ListTaskByStatus(ctx, 0, c.limit, domain.WAITING.ToUint8())
		if err != nil {
			return fmt.Errorf("获取任务列表失败: %w", err)
		}

		// 并行启动任务
		for _, task := range tasks {
			go func(t domain.Task) {
				err = c.start(ctx, t.ProcessInstId, t.CurrentNodeId) // 重新声明局部 err
				if err != nil {
					c.logger.Error("自动化任务启动失败", elog.FieldErr(err), elog.Int64("task_id", t.Id))
				}
			}(task)
		}

		if len(tasks) < int(c.limit) {
			break
		}

		if c.limit >= total {
			break
		}
	}
	return nil
}

func (c *StartTaskJob) start(ctx context.Context, processInstId int, currentNodeId string) error {
	strategy, er := retry.NewExponentialBackoffRetryStrategy(c.initialInterval, c.maxInterval, c.maxRetries)
	if er != nil {
		return er
	}

	var err error
	for {
		d, ok := strategy.Next()
		if !ok {
			c.logger.Warn("处理执行任务超过最大重试次数",
				elog.Any("processInstId", processInstId),
				elog.Any("currentNodeId", currentNodeId),
			)
			return fmt.Errorf("超过最大重试次数")
		}

		err = c.svc.StartTask(ctx, processInstId, currentNodeId)
		if err != nil {
			time.Sleep(d)
			continue
		}

		return nil
	}
}
