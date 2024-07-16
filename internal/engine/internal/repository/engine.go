package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/engine/internal/repository/dao"
)

type ProcessEngineRepository interface {
	CountTodo(ctx context.Context, userId, processName string) (int64, error)
}

type processEngineRepository struct {
	engineDao dao.ProcessEngineDAO
}

func (repo *processEngineRepository) CountTodo(ctx context.Context, userId, processName string) (int64, error) {
	return repo.engineDao.TodoCount(ctx, userId, processName)
}

func NewProcessEngineRepository(engineDao dao.ProcessEngineDAO) ProcessEngineRepository {
	return &processEngineRepository{
		engineDao: engineDao,
	}
}
