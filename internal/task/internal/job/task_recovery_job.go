package job

import (
	"context"
	"fmt"
	"sync"

	"time"

	"github.com/Duke1616/ecmdb/internal/task/domain"
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

		now := time.Now().UnixMilli()
		for _, t := range tasks {
			// 定时任务判定：如果当前时间还没到计划开始时间，跳过恢复（等待 StartTaskJob 正常触发）
			if t.IsTiming && now < t.ScheduledTime {
				continue
			}

			// 普通任务判定：给 1 分钟的「派发宽限期」。
			// 只有 Utime 超过 1 分钟的任务才被视为「卡在 SCHEDULED 状态」需要补发
			if !t.IsTiming && now < t.Utime+60*1000 {
				continue
			}

			wg.Add(1)
			go func(task domain.Task) {
				defer wg.Done()
				// 调用 service 层的自动补发逻辑，带计数上限控制
				_ = c.svc.AutoRetryTask(ctx, task.Id)
			}(t)
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
