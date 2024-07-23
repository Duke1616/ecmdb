package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository/dao"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, req domain.Task) (int64, error)
}

type taskRepository struct {
	dao dao.TaskDAO
}

func (repo *taskRepository) CreateTask(ctx context.Context, req domain.Task) (int64, error) {
	return repo.dao.CreateTask(ctx, repo.toEntity(req))
}

func NewTaskRepository(dao dao.TaskDAO) TaskRepository {
	return &taskRepository{
		dao: dao,
	}
}

func (repo *taskRepository) toEntity(req domain.Task) dao.Task {
	return dao.Task{
		ProcessInstId: req.ProcessInstId,
		CodebookUid:   req.CodebookUid,
		WorkerName:    req.WorkerName,
		WorkflowId:    req.WorkflowId,
		Code:          req.Code,
		Topic:         req.Topic,
		Language:      req.Language,
	}
}
