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
	UpdateVariables(ctx context.Context, id int64, variables string) (int64, error)
	ListTask(ctx context.Context, offset, limit int64) ([]domain.Task, error)
	ListTaskByStatus(ctx context.Context, offset, limit int64, status uint8) ([]domain.Task, error)
	Total(ctx context.Context, status uint8) (int64, error)

	UpdateArgs(ctx context.Context, id int64, args map[string]interface{}) (int64, error)
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

func (repo *taskRepository) UpdateVariables(ctx context.Context, id int64, variables string) (int64, error) {
	return repo.dao.UpdateVariables(ctx, id, variables)
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
		WorkerName:      req.WorkerName,
		WorkflowId:      req.WorkflowId,
		Code:            req.Code,
		Topic:           req.Topic,
		Language:        req.Language,
		Args:            req.Args,
		Variables:       req.Variables,
		Status:          req.Status.ToUint8(),
	}
}

func (repo *taskRepository) toDomain(req dao.Task) domain.Task {
	return domain.Task{
		Id:            req.Id,
		ProcessInstId: req.ProcessInstId,
		CurrentNodeId: req.CurrentNodeId,
		OrderId:       req.OrderId,
		CodebookUid:   req.CodebookUid,
		WorkerName:    req.WorkerName,
		WorkflowId:    req.WorkflowId,
		Code:          req.Code,
		Topic:         req.Topic,
		Args:          req.Args,
		Variables:     req.Variables,
		Language:      req.Language,
		Result:        req.Result,
		Status:        domain.Status(req.Status),
	}
}
