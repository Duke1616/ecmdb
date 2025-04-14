package repository

import (
	"context"
	"errors"
	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"go.mongodb.org/mongo-driver/mongo"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, req domain.Task) (int64, error)
	FindByProcessInstId(ctx context.Context, processInstId int, nodeId string) (domain.Task, error)
	FindOrCreate(ctx context.Context, req domain.Task) (domain.Task, error)
	FindById(ctx context.Context, id int64) (domain.Task, error)
	UpdateTask(ctx context.Context, req domain.Task) (int64, error)
	UpdateTaskStatus(ctx context.Context, req domain.TaskResult) (int64, error)
	UpdateVariables(ctx context.Context, id int64, variables []domain.Variables) (int64, error)
	ListTask(ctx context.Context, offset, limit int64) ([]domain.Task, error)
	ListTaskByStatus(ctx context.Context, offset, limit int64, status uint8) ([]domain.Task, error)
	Total(ctx context.Context, status uint8) (int64, error)
	UpdateArgs(ctx context.Context, id int64, args map[string]interface{}) (int64, error)
	ListSuccessTasksByUtime(ctx context.Context, offset, limit int64, utime int64) ([]domain.Task, error)
	TotalByUtime(ctx context.Context, utime int64) (int64, error)
	FindTaskResult(ctx context.Context, instanceId int, nodeId string) (domain.Task, error)
	ListTaskByInstanceId(ctx context.Context, offset, limit int64, instanceId int) ([]domain.Task, error)
	TotalByInstanceId(ctx context.Context, instanceId int) (int64, error)
	MarkTaskAsAutoPassed(ctx context.Context, id int64) error
}

type taskRepository struct {
	dao dao.TaskDAO
}

func (repo *taskRepository) TotalByInstanceId(ctx context.Context, instanceId int) (int64, error) {
	return repo.dao.TotalByInstanceId(ctx, instanceId)
}

func (repo *taskRepository) ListTaskByInstanceId(ctx context.Context, offset, limit int64,
	instanceId int) ([]domain.Task, error) {
	ts, err := repo.dao.ListTaskByInstanceId(ctx, offset, limit, instanceId)
	return slice.Map(ts, func(idx int, src dao.Task) domain.Task {
		return repo.toDomain(src)
	}), err
}

func (repo *taskRepository) MarkTaskAsAutoPassed(ctx context.Context, id int64) error {
	return repo.dao.MarkTaskAsAutoPassed(ctx, id)
}

func (repo *taskRepository) FindTaskResult(ctx context.Context, instanceId int, nodeId string) (domain.Task, error) {
	task, err := repo.dao.FindTaskResult(ctx, instanceId, nodeId)
	return repo.toDomain(task), err
}

func (repo *taskRepository) ListSuccessTasksByUtime(ctx context.Context, offset, limit int64,
	utime int64) ([]domain.Task, error) {
	ts, err := repo.dao.ListSuccessTasksByUtime(ctx, offset, limit, utime)
	return slice.Map(ts, func(idx int, src dao.Task) domain.Task {
		return repo.toDomain(src)
	}), err
}

func (repo *taskRepository) TotalByUtime(ctx context.Context, utime int64) (int64, error) {
	return repo.dao.TotalByUtime(ctx, utime)
}

func (repo *taskRepository) ListTask(ctx context.Context, offset, limit int64) ([]domain.Task, error) {
	ts, err := repo.dao.ListTask(ctx, offset, limit)
	return slice.Map(ts, func(idx int, src dao.Task) domain.Task {
		return repo.toDomain(src)
	}), err
}

func (repo *taskRepository) UpdateVariables(ctx context.Context, id int64, variables []domain.Variables) (int64, error) {
	return repo.dao.UpdateVariables(ctx, id, slice.Map(variables, func(idx int, src domain.Variables) dao.Variables {
		return dao.Variables{
			Key:    src.Key,
			Value:  src.Value,
			Secret: src.Secret,
		}
	}))
}

func (repo *taskRepository) FindById(ctx context.Context, id int64) (domain.Task, error) {
	task, err := repo.dao.FindById(ctx, id)
	return repo.toDomain(task), err
}

func (repo *taskRepository) UpdateArgs(ctx context.Context, id int64, args map[string]interface{}) (int64, error) {
	return repo.dao.UpdateArgs(ctx, id, args)
}

func (repo *taskRepository) FindOrCreate(ctx context.Context, req domain.Task) (domain.Task, error) {
	// 先创建任务、以防后续失败，导致无法溯源
	task, err := repo.dao.FindByProcessInstId(ctx, req.ProcessInstId, req.CurrentNodeId)
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return repo.toDomain(task), nil
	}

	taskId, err := repo.dao.CreateTask(ctx, repo.toEntity(req))
	if err != nil {
		return domain.Task{}, err
	}

	task.Id = taskId
	return repo.toDomain(task), err
}

func (repo *taskRepository) FindByProcessInstId(ctx context.Context, processInstId int, nodeId string) (
	domain.Task, error) {
	task, err := repo.dao.FindByProcessInstId(ctx, processInstId, nodeId)
	return repo.toDomain(task), err
}

func (repo *taskRepository) UpdateTask(ctx context.Context, req domain.Task) (int64, error) {
	return repo.dao.UpdateTask(ctx, repo.toEntity(req))
}

func (repo *taskRepository) ListTaskByStatus(ctx context.Context, offset, limit int64, status uint8) ([]domain.Task, error) {
	ts, err := repo.dao.ListTaskByStatus(ctx, offset, limit, status)
	return slice.Map(ts, func(idx int, src dao.Task) domain.Task {
		return repo.toDomain(src)
	}), err
}

func (repo *taskRepository) Total(ctx context.Context, status uint8) (int64, error) {
	return repo.dao.Count(ctx, status)
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
		Id:              req.Id,
		Result:          req.Result,
		WantResult:      req.WantResult,
		Status:          req.Status.ToUint8(),
		TriggerPosition: req.TriggerPosition,
	}
}

func (repo *taskRepository) toEntity(req domain.Task) dao.Task {
	return dao.Task{
		Id:              req.Id,
		ProcessInstId:   req.ProcessInstId,
		TriggerPosition: req.TriggerPosition,
		CurrentNodeId:   req.CurrentNodeId,
		OrderId:         req.OrderId,
		CodebookUid:     req.CodebookUid,
		CodebookName:    req.CodebookName,
		WorkerName:      req.WorkerName,
		WorkflowId:      req.WorkflowId,
		Code:            req.Code,
		Topic:           req.Topic,
		Language:        req.Language,
		Args:            req.Args,
		IsTiming:        req.IsTiming,
		Timing: dao.Timing{
			Stime:    req.Timing.Stime,
			Unit:     req.Timing.Unit.ToUint8(),
			Quantity: req.Timing.Quantity,
		},
		Variables: slice.Map(req.Variables, func(idx int, src domain.Variables) dao.Variables {
			return dao.Variables{
				Key:    src.Key,
				Value:  src.Value,
				Secret: src.Secret,
			}
		}),
		Status: req.Status.ToUint8(),
	}
}

func (repo *taskRepository) toDomain(req dao.Task) domain.Task {
	return domain.Task{
		Id:            req.Id,
		ProcessInstId: req.ProcessInstId,
		CurrentNodeId: req.CurrentNodeId,
		OrderId:       req.OrderId,
		CodebookUid:   req.CodebookUid,
		CodebookName:  req.CodebookName,
		WorkerName:    req.WorkerName,
		WorkflowId:    req.WorkflowId,
		Code:          req.Code,
		Topic:         req.Topic,
		Args:          req.Args,
		IsTiming:      req.IsTiming,
		Timing: domain.Timing{
			Stime:    req.Timing.Stime,
			Unit:     domain.Unit(req.Timing.Unit),
			Quantity: req.Timing.Quantity,
		},
		Variables: slice.Map(req.Variables, func(idx int, src dao.Variables) domain.Variables {
			return domain.Variables{
				Key:    src.Key,
				Value:  src.Value,
				Secret: src.Secret,
			}
		}),
		Language:   req.Language,
		Utime:      req.Utime,
		Result:     req.Result,
		WantResult: req.WantResult,
		Status:     domain.Status(req.Status),
	}
}
