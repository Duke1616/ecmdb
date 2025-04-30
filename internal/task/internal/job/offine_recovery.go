package job

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/gotomicro/ego/core/elog"
	"sync"
	"time"
)

type RecoveryTaskJob struct {
	svc     service.Service
	execSvc service.ExecService
	cronSvc service.Cronjob
	logger  *elog.Component
	limit   int64
}

func NewRecoveryTaskJob(svc service.Service, execSvc service.ExecService,
	cronSvc service.Cronjob, limit int64) *RecoveryTaskJob {
	return &RecoveryTaskJob{
		svc:     svc,
		execSvc: execSvc,
		cronSvc: cronSvc,
		logger:  elog.DefaultLogger,
		limit:   limit,
	}
}

func (c *RecoveryTaskJob) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	for {
		tasks, total, err := c.svc.ListTaskByStatus(ctx, 0, c.limit, domain.TIMING.ToUint8())
		if err != nil {
			return fmt.Errorf("recovery 获取任务列表失败: %w", err)
		}
		c.logger.Info("recovery task job start", elog.Int64("total", total))

		for _, task := range tasks {
			wg.Add(1)
			go func(task domain.Task) {
				defer wg.Done()
				now := time.Now()
				if time.UnixMilli(task.Timing.Stime).Before(now) || time.UnixMilli(task.Timing.Stime).Equal(now) {
					_ = c.execSvc.Execute(ctx, task)
				} else {
					_ = c.cronSvc.Create(ctx, task)
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

	wg.Wait()
	return nil
}
