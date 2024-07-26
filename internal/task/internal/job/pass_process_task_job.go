package job

import (
	"context"
	"errors"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/gotomicro/ego/task/ecron"
	"gorm.io/gorm"
	"time"
)

var _ ecron.NamedJob = (*PassProcessTaskJob)(nil)

type PassProcessTaskJob struct {
	svc       service.Service
	engineSvc engine.Service
	minutes   int64
	seconds   int64
	limit     int64
}

func NewPassProcessTaskJob(svc service.Service, engineSvc engine.Service, minutes, seconds int64,
	limit int64) *PassProcessTaskJob {
	return &PassProcessTaskJob{
		svc:       svc,
		engineSvc: engineSvc,
		minutes:   minutes,
		seconds:   seconds,
		limit:     limit,
	}
}

func (c *PassProcessTaskJob) Name() string {
	return "PassProcessTask"
}

func (c *PassProcessTaskJob) Run(ctx context.Context) error {
	// 我只处理近一个小时的数据，进行比对
	ctime := time.Now().Add(time.Duration(-c.minutes)*time.Minute + time.Duration(-c.seconds)*time.Second).UnixMilli()
	for {
		// 获取执行任务
		tasks, total, err := c.svc.ListTasksByCtime(ctx, 0, c.limit, ctime)
		if err != nil {
			return fmt.Errorf("查询执行任务失败: %w", err)
		}

		// 遍历
		for _, task := range tasks {
			// 获取自动化步骤
			exist, er := c.engineSvc.GetAutomationTask(ctx, task.CurrentNodeId, task.ProcessInstId)
			if errors.Is(er, gorm.ErrRecordNotFound) {
				continue
			}

			// 如果返回有值，则通过
			if exist.TaskID != 0 {
				err = c.engineSvc.Pass(ctx, exist.TaskID, "任务执行完成")
				if err != nil {
					return fmt.Errorf("通过自动化节点失败: %w", err)
				}
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
