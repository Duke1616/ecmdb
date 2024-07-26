package ioc

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/task/ecron"
	"time"
)

func initCronJobs(tJob *task.StartTaskJob) []*ecron.Component {
	loc, _ := time.LoadLocation("Asia/Shanghai")

	return []*ecron.Component{
		ecron.DefaultContainer().Build(
			ecron.WithJob(funcJobWrapper(tJob)),
			ecron.WithSeconds(),
			ecron.WithSpec("*/10 * * * * *"),
			ecron.WithLocation(loc),
		),
	}
}

func funcJobWrapper(job ecron.NamedJob) ecron.FuncJob {
	name := job.Name()
	return func(ctx context.Context) error {
		start := time.Now()
		elog.DefaultLogger.Debug("开始运行",
			elog.String("cronjob", name))
		err := job.Run(ctx)
		if err != nil {
			elog.DefaultLogger.Error("执行失败",
				elog.FieldErr(err),
				elog.String("cronjob", name))
			return err
		}
		duration := time.Since(start)
		elog.DefaultLogger.Debug("结束运行",
			elog.String("cronjob", name),
			elog.FieldKey("运行时间"),
			elog.FieldCost(duration))
		return nil
	}
}
