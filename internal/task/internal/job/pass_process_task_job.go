package job

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/task/ecron"
	"gorm.io/gorm"
)

var _ ecron.NamedJob = (*PassProcessTaskJob)(nil)

type PassProcessTaskJob struct {
	logger    *elog.Component
	svc       service.Service
	taskIds   []string
	engineSvc engine.Service
	minutes   int64
	seconds   int64
	offset    int64
	limit     int64
}

func NewPassProcessTaskJob(svc service.Service, engineSvc engine.Service, minutes, seconds int64,
	limit int64) *PassProcessTaskJob {
	return &PassProcessTaskJob{
		taskIds:   make([]string, 0),
		logger:    elog.DefaultLogger,
		svc:       svc,
		engineSvc: engineSvc,
		minutes:   minutes,
		seconds:   seconds,
		offset:    0,
		limit:     limit,
	}
}

func (c *PassProcessTaskJob) Name() string {
	return "PassProcessTask"
}

func (c *PassProcessTaskJob) Run(ctx context.Context) error {
	// 10 分钟前置延迟、自动结束自动化任务节点
	utime := time.Now().Add(time.Duration(-c.minutes)*time.Minute + time.Duration(-c.seconds)*time.Second).UnixMilli()
	for {
		// 获取执行任务
		tasks, total, err := c.svc.ListSuccessTasksByUtime(ctx, c.offset, c.limit, utime)
		if err != nil {
			return fmt.Errorf("查询执行任务失败: %w", err)
		}

		// 遍历
		for _, task := range tasks {
			c.logger.Info("任务开启自动通过逻辑", elog.Int64("id", task.Id))
			// 获取自动化步骤
			mt := model.Task{}
			mt, err = c.engineSvc.GetAutomationTask(ctx, task.CurrentNodeId, task.ProcessInstId)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}

			if err != nil {
				c.logger.Error("获取自动化任务失败", elog.FieldErr(err))
			}

			// 如果返回有值，则通过
			if mt.TaskID != 0 {
				err = c.engineSvc.Pass(ctx, mt.TaskID, "任务执行完成")
				if err != nil {
					return fmt.Errorf("通过自动化节点失败: %w", err)
				}
			}

			// 标记数据已经成功
			err = c.svc.MarkTaskAsAutoPassed(ctx, task.Id)
			if err != nil {
				c.logger.Error("数据标记失败", elog.FieldErr(err))
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
