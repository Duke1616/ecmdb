package dispatch

import (
	"context"

	taskv1 "github.com/Duke1616/ecmdb/api/proto/gen/etask/task/v1"
	"github.com/Duke1616/ecmdb/internal/task/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
)

type TaskDispatcher interface {
	Dispatch(ctx context.Context, task domain.Task) error
}

type taskDispatcher struct {
	kafkaSvc   TaskDispatcher
	executeSvc TaskDispatcher
}

func NewTaskDispatcher(workerSvc worker.Service, grpcClient taskv1.TaskServiceClient,
	repo repository.TaskRepository, crypto cryptox.Crypto) TaskDispatcher {
	return &taskDispatcher{
		kafkaSvc:   NewKafkaService(workerSvc, crypto),
		executeSvc: NewExecuteService(grpcClient, repo, crypto),
	}
}

func (d *taskDispatcher) Dispatch(ctx context.Context, task domain.Task) error {
	if task.RunMode == domain.RunModeExecute {
		return d.executeSvc.Dispatch(ctx, task)
	}

	return d.kafkaSvc.Dispatch(ctx, task)
}
