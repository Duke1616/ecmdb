package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, req domain.Task) (int64, error)
	UpdateTaskStatus(ctx context.Context, req domain.TaskResult) (int64, error)
	ListTask(ctx context.Context, offset, limit int64) ([]domain.Task, error)
	Total(ctx context.Context) (int64, error)
}

type taskRepository struct {
	dao dao.TaskDAO
}

func (repo *taskRepository) ListTask(ctx context.Context, offset, limit int64) ([]domain.Task, error) {
	ts, err := repo.dao.ListTask(ctx, offset, limit)
	return slice.Map(ts, func(idx int, src dao.Task) domain.Task {
		return repo.toDomain(src)
	}), err
}

func (repo *taskRepository) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func (repo *taskRepository) UpdateTaskStatus(ctx context.Context, req domain.TaskResult) (int64, error) {
	return repo.dao.UpdateTaskStatus(ctx, repo.toUpdateEntity(req))
}

func (repo *taskRepository) CreateTask(ctx context.Context, req domain.Task) (int64, error) {
	return repo.dao.CreateTask(ctx, repo.toEntity(req))
}

func NewTaskRepository(dao dao.TaskDAO) TaskRepository {
	return &taskRepository{
		dao: dao,
	}
}

func (repo *taskRepository) toUpdateEntity(req domain.TaskResult) dao.Task {
	return dao.Task{
		Id:     req.Id,
		Result: req.Result,
		Status: req.Status.ToUint8(),
	}
}

func (repo *taskRepository) toEntity(req domain.Task) dao.Task {
	return dao.Task{
		ProcessInstId: req.ProcessInstId,
		OrderId:       req.OrderId,
		CodebookUid:   req.CodebookUid,
		WorkerName:    req.WorkerName,
		WorkflowId:    req.WorkflowId,
		Code:          req.Code,
		Topic:         req.Topic,
		Language:      req.Language,
		Args:          req.Args,
		Status:        req.Status.ToUint8(),
	}
}

func (repo *taskRepository) toDomain(req dao.Task) domain.Task {
	return domain.Task{
		Id:            req.Id,
		ProcessInstId: req.ProcessInstId,
		OrderId:       req.OrderId,
		CodebookUid:   req.CodebookUid,
		WorkerName:    req.WorkerName,
		WorkflowId:    req.WorkflowId,
		Code:          req.Code,
		Topic:         req.Topic,
		Args:          req.Args,
		Language:      req.Language,
		Result:        req.Result,
		Status:        domain.Status(req.Status),
	}
}
