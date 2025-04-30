package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/gotomicro/ego/core/elog"
	"github.com/robfig/cron/v3"
	"time"
)

type Cronjob interface {
	Create(ctx context.Context, task domain.Task) error
}

type cronjob struct {
	execService ExecService
	logger      *elog.Component
	cron        *cron.Cron
}

func NewCronjob(execService ExecService) Cronjob {
	return &cronjob{
		execService: execService,
		logger:      elog.DefaultLogger,
		cron:        cron.New(cron.WithSeconds()),
	}
}

func (c *cronjob) Create(ctx context.Context, task domain.Task) error {
	// jobTime 秒、分、时、日、月
	jobTime := time.UnixMilli(task.Timing.Stime).Format("05 04 15 02 01 *")
	_, err := c.cron.AddFunc(jobTime, func() {
		if err := c.execService.Execute(ctx, task); err != nil {
			c.logger.Error("Task execution failed", elog.FieldErr(err))
			return
		}
		c.logger.Info("task execution finished successfully", elog.Int64("task_id", task.Id))
		return
	})
	if err != nil {
		return err
	}

	c.logger.Info("starting task execution", elog.Any("time", jobTime), elog.Int64("task_id", task.Id))
	c.cron.Start()
	return nil
}
