package dispatch

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"github.com/robfig/cron/v3"
)

type kafkaService struct {
	workerSvc worker.Service
	crypto    cryptox.Crypto
	cron      *cron.Cron
	logger    *elog.Component
}

func NewKafkaService(workerSvc worker.Service, crypto cryptox.Crypto) TaskDispatcher {
	k := &kafkaService{
		workerSvc: workerSvc,
		crypto:    crypto,
		cron:      cron.New(cron.WithSeconds()),
		logger:    elog.DefaultLogger,
	}
	k.cron.Start()
	return k
}

func (e *kafkaService) Dispatch(ctx context.Context, task domain.Task) error {
	if task.IsTiming && task.Timing.Stime > time.Now().UnixMilli() {
		return e.scheduleTimingTask(ctx, task)
	}

	return e.immediateDispatch(ctx, task)
}

func (e *kafkaService) immediateDispatch(ctx context.Context, task domain.Task) error {
	return e.workerSvc.Execute(ctx, worker.Execute{
		TaskId:    task.Id,
		Topic:     task.Topic,
		Code:      task.Code,
		Language:  task.Language,
		Args:      task.Args,
		Variables: e.decryptVariables(task.Variables),
	})
}

func (e *kafkaService) scheduleTimingTask(ctx context.Context, task domain.Task) error {
	jobTime := time.UnixMilli(task.Timing.Stime).Format("05 04 15 02 01 ?")
	_, err := e.cron.AddFunc(jobTime, func() {
		if err := e.immediateDispatch(ctx, task); err != nil {
			e.logger.Error("Worker 定时任务执行失败", elog.FieldErr(err), elog.Int64("taskId", task.Id))
			return
		}
		e.logger.Info("Worker 定时任务执行成功", elog.Int64("taskId", task.Id))
	})
	if err != nil {
		return err
	}

	e.logger.Info("Worker 任务开始定时调度", elog.Any("time", jobTime), elog.Int64("taskId", task.Id))
	return nil
}

// decryptVariables 处理变量，进行解密
func (e *kafkaService) decryptVariables(req []domain.Variables) string {
	variables := slice.Map(req, func(idx int, src domain.Variables) domain.Variables {
		if src.Secret {
			val, er := e.crypto.Decrypt(src.Value)
			if er != nil {
				return domain.Variables{}
			}

			return domain.Variables{
				Key:    src.Key,
				Value:  val,
				Secret: src.Secret,
			}
		}

		return domain.Variables{
			Key:    src.Key,
			Value:  src.Value,
			Secret: src.Secret,
		}
	})

	jsonVar, _ := json.Marshal(variables)
	return string(jsonVar)
}
