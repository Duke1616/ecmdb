package job

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/gotomicro/ego/task/ecron"
)

var _ ecron.NamedJob = (*StartTaskJob)(nil)

type StartTaskJob struct {
	svc   service.Service
	limit int64
}

func NewStartTaskJob(svc service.Service, limit int64) *StartTaskJob {
	return &StartTaskJob{
		svc:   svc,
		limit: limit,
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

		for _, task := range tasks {
			err = c.svc.StartTask(ctx, task.ProcessInstId, task.CurrentNodeId)

			if err != nil {
				return fmt.Errorf("启动任务失败: %w", err)
			}
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
