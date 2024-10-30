package task

import (
	"context"
	"errors"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/ecodeclub/ekit/retry"
	"github.com/gotomicro/ego/core/elog"
	"time"
)

type ExponentialBackoffTaskFetcher struct {
	engineSvc       engine.Service
	initialInterval time.Duration
	maxInterval     time.Duration
	maxRetries      int32
	logger          *elog.Component
}

func NewExponentialBackoffTaskFetcher(engineSvc engine.Service) *ExponentialBackoffTaskFetcher {
	return &ExponentialBackoffTaskFetcher{
		engineSvc:       engineSvc,
		initialInterval: 5 * time.Second,
		maxInterval:     15 * time.Second,
		maxRetries:      3,
		logger:          elog.DefaultLogger,
	}

}

func (f *ExponentialBackoffTaskFetcher) FetchTasksWithRetry(ctx context.Context, instanceId int, userIDs []string) ([]model.Task, error) {
	strategy, err := retry.NewExponentialBackoffRetryStrategy(f.initialInterval, f.maxInterval, f.maxRetries)
	if err != nil {
		return nil, err
	}

	for {
		d, ok := strategy.Next()
		if !ok {
			f.logger.Error("超过最大重试次数", elog.Any("instId", instanceId))
			return nil, errors.New("max retries reached")
		}

		tasks, er := f.engineSvc.GetTasksByInstUsers(ctx, instanceId, userIDs)
		if er != nil || len(tasks) == 0 {
			time.Sleep(d)
			continue
		}

		return tasks, nil
	}
}
