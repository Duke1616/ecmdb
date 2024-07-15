package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/engine/internal/repository/dao"
)

type TaskRepository interface {
	CountTodo(ctx context.Context, userId, processName string) (int64, error)
}

type taskRepository struct {
	engineDao dao.ProcessEngineDAO
}

func (repo *taskRepository) CountTodo(ctx context.Context, userId, processName string) (int64, error) {
	return repo.engineDao.TodoCount(ctx, userId, processName)
}

func NewTaskRepository(engineDao dao.ProcessEngineDAO) TaskRepository {
	return &taskRepository{
		engineDao: engineDao,
	}
}
