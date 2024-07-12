package repository

import (
	"github.com/Duke1616/ecmdb/internal/task/repository/dao"
)

type TaskRepository interface {
}

type taskRepository struct {
	dao dao.TaskDAO
}

func NewTaskRepository(dao dao.TaskDAO) TaskRepository {
	return &taskRepository{
		dao: dao,
	}
}
