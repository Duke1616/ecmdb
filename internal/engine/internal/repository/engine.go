package repository

import (
	"context"
	"github.com/Bunny3th/easy-workflow/workflow/database"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/engine/internal/domain"
	"github.com/Duke1616/ecmdb/internal/engine/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type ProcessEngineRepository interface {
	ListTodoList(userId, processName string, sortByAse bool, offset, limit int) (
		[]domain.Instance, error)
	CountTodo(ctx context.Context, userId, processName string) (int64, error)
	CountStartUser(ctx context.Context, userId, processName string) (int64, error)
	ListStartUser(ctx context.Context, userId, processName string, offset, limit int) (
		[]domain.Instance, error)
}

type processEngineRepository struct {
	engineDao dao.ProcessEngineDAO
}

func (repo *processEngineRepository) ListTodoList(userId, processName string, sortByAse bool, offset, limit int) ([]domain.Instance, error) {
	ts, err := engine.GetTaskToDoList(userId, processName, sortByAse, offset, limit)
	return slice.Map(ts, func(idx int, src model.Task) domain.Instance {
		return repo.toDomainByTask(src)
	}), err
}

func (repo *processEngineRepository) ListStartUser(ctx context.Context, userId, processName string, offset,
	limit int) ([]domain.Instance, error) {
	ts, err := repo.engineDao.ListStartUser(ctx, userId, processName, offset, limit)
	return slice.Map(ts, func(idx int, src dao.Instance) domain.Instance {
		return repo.toDomainByInstance(src)
	}), err
}

func (repo *processEngineRepository) CountTodo(ctx context.Context, userId, processName string) (int64, error) {
	return repo.engineDao.CountTodo(ctx, userId, processName)
}

func (repo *processEngineRepository) CountStartUser(ctx context.Context, userId, processName string) (int64, error) {
	return repo.engineDao.CountStartUser(ctx, userId, processName)
}

func NewProcessEngineRepository(engineDao dao.ProcessEngineDAO) ProcessEngineRepository {
	return &processEngineRepository{
		engineDao: engineDao,
	}
}

func (repo *processEngineRepository) toDomainByInstance(req dao.Instance) domain.Instance {
	return domain.Instance{
		TaskID:          req.TaskID,
		ProcInstID:      req.ProcInstID,
		ProcVersion:     req.ProcVersion,
		ProcID:          req.ProcID,
		ProcName:        req.ProcName,
		Status:          req.Status,
		CreateTime:      (*database.LocalTime)(req.CreateTime),
		CurrentNodeID:   req.CurrentNodeID,
		CurrentNodeName: req.CurrentNodeName,
		BusinessID:      req.BusinessID,
		ApprovedBy:      []string{req.ApprovedBy},
		Starter:         req.Starter,
	}
}

func (repo *processEngineRepository) toDomainByTask(req model.Task) domain.Instance {
	return domain.Instance{
		TaskID:          req.TaskID,
		ProcInstID:      req.ProcInstID,
		ProcID:          req.ProcID,
		ProcName:        req.ProcName,
		BusinessID:      req.BusinessID,
		Starter:         req.Starter,
		CurrentNodeID:   req.NodeID,
		CurrentNodeName: req.NodeName,
		CreateTime:      (*database.LocalTime)(req.CreateTime),
		ApprovedBy:      []string{req.UserID},
		Status:          req.Status,
	}
}
