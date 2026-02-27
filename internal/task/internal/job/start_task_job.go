package job

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/task/domain"
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
		tasks, err := c.svc.ListReadyTasks(ctx, c.limit)
		if err != nil {
			return fmt.Errorf("捞取就绪任务失败: %w", err)
		}

		// 无任务则退出本次循环
		if len(tasks) == 0 {
			break
		}

		// 顺序/并发由具体业务决定，这里采用并发启动以提升吞吐
		for _, task := range tasks {
			go func(t domain.Task) {
				// 独立 Context 避免受 Job 周期 Context 过期影响
				err = c.start(context.Background(), t.Id)
				if err != nil {
					c.logger.Error("就绪任务启动失败", elog.FieldErr(err), elog.Int64("taskId", t.Id))
				}
			}(task)
		}

		// 如果捞取的任务少于 limit，说明当前周期已处理完
		if len(tasks) < int(c.limit) {
			break
		}

		// 避免过于频繁的循环，稍微喘息
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (c *StartTaskJob) start(ctx context.Context, id int64) error {
	strategy, er := retry.NewExponentialBackoffRetryStrategy(c.initialInterval, c.maxInterval, c.maxRetries)
	if er != nil {
		return er
	}

	var err error
	for {
		d, ok := strategy.Next()
		if !ok {
			c.logger.Warn("启动就绪任务超过最大重试次数",
				elog.Int64("taskId", id),
			)
			return fmt.Errorf("超过最大重试次数")
		}

		err = c.svc.StartTask(ctx, id)
		if err != nil {
			c.logger.Warn("任务启动重试中", elog.Int64("taskId", id), elog.FieldErr(err))
			time.Sleep(d)
			continue
		}

		return nil
	}
}
