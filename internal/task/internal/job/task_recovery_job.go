package job

import (
	"context"
	"fmt"
	"sync"

	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/gotomicro/ego/core/elog"
)

type TaskRecoveryJob struct {
	svc    service.Service
	logger *elog.Component
	limit  int64
}

func NewTaskRecoveryJob(svc service.Service, limit int64) *TaskRecoveryJob {
	return &TaskRecoveryJob{
		svc:    svc,
		logger: elog.DefaultLogger,
		limit:  limit,
	}
}

func (c *TaskRecoveryJob) Name() string {
	return "TaskRecoveryJob"
}

func (c *TaskRecoveryJob) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	for {
		tasks, total, err := c.svc.ListTaskByStatus(ctx, 0, c.limit, domain.SCHEDULED.ToUint8())
		if err != nil {
			return fmt.Errorf("recovery 获取任务列表失败: %w", err)
		}
		c.logger.Info("recovery task job start", elog.Int64("total", total))

		for _, task := range tasks {
			wg.Add(1)
			go func(task domain.Task) {
				defer wg.Done()
				// 调用 service 层的自动补发逻辑，带计数上限控制
				_ = c.svc.AutoRetryTask(ctx, task.Id)
			}(task)
		}

		if len(tasks) < int(c.limit) {
			break
		}

		if c.limit >= total {
			break
		}
	}

	wg.Wait()
	return nil
}
