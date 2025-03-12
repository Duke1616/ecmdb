package service

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/gotomicro/ego/task/ecron"
	"time"
)

type Cronjob interface {
	Add(ctx context.Context, task domain.Task) error
}

type cronjob struct {
	execService ExecService
}

func NewCronjob(execService ExecService) Cronjob {
	return &cronjob{
		execService: execService,
	}
}

func (c *cronjob) Add(ctx context.Context, task domain.Task) error {
	loc, _ := time.LoadLocation("Asia/Shanghai")

	job := ecron.DefaultContainer().Build(
		ecron.WithJob(func(ctx context.Context) error {
			return c.execService.Execute(ctx, task)
		}),
		ecron.WithSeconds(),
		ecron.WithSpec(expr(task.Timing.Stime)),
		ecron.WithLocation(loc),
	)

	return job.Start()
}

func expr(startTime int64) string {
	targetTime := time.Unix(startTime, 0)
	return fmt.Sprintf("* %d %d %d * *", targetTime.Minute(), targetTime.Hour(), targetTime.Day())
}
