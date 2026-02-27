package ioc

import (
	"context"
	"time"

	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/task/ecron"
)

func initCronJobs(tJob *task.StartTaskJob, pJob *task.PassProcessTaskJob, sJob *task.TaskExecutionSyncJob, rJob *task.TaskRecoveryJob) []*ecron.Component {
	loc, _ := time.LoadLocation("Asia/Shanghai")

	return []*ecron.Component{
		ecron.DefaultContainer().Build(
			ecron.WithJob(funcJobWrapper(tJob)),
			ecron.WithSeconds(),
			ecron.WithSpec("*/10 * * * * *"),
			ecron.WithLocation(loc),
		),
		ecron.DefaultContainer().Build(
			ecron.WithJob(funcJobWrapper(pJob)),
			ecron.WithSeconds(),
			ecron.WithSpec("*/10 * * * * *"),
			ecron.WithLocation(loc),
		),
		ecron.DefaultContainer().Build(
			ecron.WithJob(funcJobWrapper(sJob)),
			ecron.WithSeconds(),
			ecron.WithSpec("*/10 * * * * *"),
			ecron.WithLocation(loc),
		),
		// NOTE: offine_recovery 负责扫描 SCHEDULED 状态的任务并重试，
		// 间隔不宜过短（避免与正常派发冲突），30s 兼顾响应速度与系统压力。
		// 内存锁（scheduler）保证同一任务不会被并发重复派发。
		ecron.DefaultContainer().Build(
			ecron.WithJob(funcJobWrapper(rJob)),
			ecron.WithSeconds(),
			ecron.WithSpec("*/30 * * * * *"),
			ecron.WithLocation(loc),
		),
	}
}

// funcJobWrapper 将 NamedJob 包装为 ecron.FuncJob，同时注入 Info 级别的开始/结束日志。
// 因此在此处以 Info 级别打印含 job 名的日志，便于生产环境运维清晰识别每次调度。
func funcJobWrapper(job ecron.NamedJob) ecron.FuncJob {
	name := job.Name()
	return func(ctx context.Context) error {
		start := time.Now()
		elog.DefaultLogger.Debug("cronjob 开始", elog.String("job", name))

		err := job.Run(ctx)
		if err != nil {
			elog.DefaultLogger.Error("cronjob 执行失败",
				elog.String("job", name),
				elog.FieldErr(err),
			)
			return err
		}

		elog.DefaultLogger.Debug("cronjob 结束",
			elog.String("job", name),
			elog.FieldCost(time.Since(start)),
		)
		return nil
	}
}
