package repository

import (
	"context"
	"errors"

	"github.com/Duke1616/ecmdb/internal/task/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"go.mongodb.org/mongo-driver/mongo"
)

type TaskRepository interface {
	// CreateTask 创建新的任务领域模型实例
	CreateTask(ctx context.Context, req domain.Task) (domain.Task, error)

	// FindByProcessInstId 根据流程实例与节点提取对应任务节点模型
	FindByProcessInstId(ctx context.Context, processInstId int, nodeId string) (domain.Task, error)

	// FindOrCreate 查询指定任务，若不存在则创建缺省记录避免丢数
	FindOrCreate(ctx context.Context, req domain.Task) (domain.Task, error)

	// FindById 依据标识 ID 获取领域模型
	FindById(ctx context.Context, id int64) (domain.Task, error)

	// UpdateTask 更新任务数据并落库
	UpdateTask(ctx context.Context, req domain.Task) (int64, error)

	// UpdateTaskStatus 同步更新当前任务的状态和最终执行结果
	UpdateTaskStatus(ctx context.Context, req domain.TaskResult) (int64, error)

	// UpdateVariables 对任务环境变量字段进行修补变更
	UpdateVariables(ctx context.Context, id int64, variables []domain.Variables) (int64, error)

	// ListTask 拉取分页的全部任务集合
	ListTask(ctx context.Context, offset, limit int64) ([]domain.Task, error)

	// ListTaskByStatus 拉取某具体状态项的分页集合
	ListTaskByStatus(ctx context.Context, offset, limit int64, status uint8) ([]domain.Task, error)

	// ListTaskByStatusAndKind 拉取特定状态和运行模式的分页集合
	ListTaskByStatusAndKind(ctx context.Context, offset, limit int64, status uint8, kind string) ([]domain.Task, error)

	// Total 计算任务集合总量
	Total(ctx context.Context, status uint8) (int64, error)

	// TotalByStatusAndKind 统计特定状态和运行模式的任务总数
	TotalByStatusAndKind(ctx context.Context, status uint8, kind string) (int64, error)

	// UpdateArgs 动态更新调度派发的额外业务配置参数
	UpdateArgs(ctx context.Context, id int64, args map[string]interface{}) (int64, error)

	// ListSuccessTasksByUtime 拉取满足特定更新时间游标且尚未读取跳过的已成行成功任务池
	ListSuccessTasksByUtime(ctx context.Context, offset, limit int64, utime int64) ([]domain.Task, error)

	// TotalByUtime 测算该更新时间范围下成功任务条目数量
	TotalByUtime(ctx context.Context, utime int64) (int64, error)

	// FindTaskResult 请求匹配的工作流特定节点的成活结果镜像封装
	FindTaskResult(ctx context.Context, instanceId int, nodeId string) (domain.Task, error)

	// ListReadyTasks 捞取已经准备好可以执行的 WAITING 任务（定时任务需满足执行时间）
	ListReadyTasks(ctx context.Context, limit int64) ([]domain.Task, error)

	// ListTaskByInstanceId 基于给定的流程树全量列出所有已触发分配的从属任务实例群落
	ListTaskByInstanceId(ctx context.Context, offset, limit int64, instanceId int) ([]domain.Task, error)

	// TotalByInstanceId 查询指定工单/流程触发的工作任务实体子节点计数
	TotalByInstanceId(ctx context.Context, instanceId int) (int64, error)

	// MarkTaskAsAutoPassed 将执行完成的任务打上流程驱动完毕可直接跳过的标记
	MarkTaskAsAutoPassed(ctx context.Context, id int64) error

	// UpdateExternalId 绑定外部分布式平台的任务 ID
	UpdateExternalId(ctx context.Context, id int64, externalId string) error
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
	if err != nil {
		return nil, err
	}

	return slice.Map(ts, func(idx int, src dao.Task) domain.Task {
		return repo.toDomain(src)
	}), nil
}

func (repo *taskRepository) ListReadyTasks(ctx context.Context, limit int64) ([]domain.Task, error) {
	ts, err := repo.dao.ListReadyTasks(ctx, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(ts, func(idx int, src dao.Task) domain.Task {
		return repo.toDomain(src)
	}), nil
}

func (repo *taskRepository) MarkTaskAsAutoPassed(ctx context.Context, id int64) error {
	return repo.dao.MarkTaskAsAutoPassed(ctx, id)
}

func (repo *taskRepository) UpdateExternalId(ctx context.Context, id int64, externalId string) error {
	return repo.dao.UpdateExternalId(ctx, id, externalId)
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

	t, err := repo.dao.CreateTask(ctx, repo.toEntity(req))
	if err != nil {
		return domain.Task{}, err
	}

	return repo.toDomain(t), nil
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

func (repo *taskRepository) ListTaskByStatusAndKind(ctx context.Context, offset, limit int64, status uint8, kind string) ([]domain.Task, error) {
	ts, err := repo.dao.ListTaskByStatusAndKind(ctx, offset, limit, status, kind)
	return slice.Map(ts, func(idx int, src dao.Task) domain.Task {
		return repo.toDomain(src)
	}), err
}

func (repo *taskRepository) Total(ctx context.Context, status uint8) (int64, error) {
	return repo.dao.Count(ctx, status)
}

func (repo *taskRepository) TotalByStatusAndKind(ctx context.Context, status uint8, kind string) (int64, error) {
	return repo.dao.CountByStatusAndKind(ctx, status, kind)
}

func (repo *taskRepository) UpdateTaskStatus(ctx context.Context, req domain.TaskResult) (int64, error) {
	return repo.dao.UpdateTaskStatus(ctx, repo.toUpdateEntity(req))
}

func (repo *taskRepository) CreateTask(ctx context.Context, req domain.Task) (domain.Task, error) {
	t, err := repo.dao.CreateTask(ctx, repo.toEntity(req))
	if err != nil {
		return domain.Task{}, err
	}

	return repo.toDomain(t), nil
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
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		RetryCount:      req.RetryCount,
	}
}

func (repo *taskRepository) toEntity(req domain.Task) dao.Task {
	t := dao.Task{
		Id:              req.Id,
		ProcessInstId:   req.ProcessInstId,
		TriggerPosition: req.TriggerPosition,
		CurrentNodeId:   req.CurrentNodeId,
		OrderId:         req.OrderId,
		CodebookUid:     req.CodebookUid,
		CodebookName:    req.CodebookName,
		WorkflowId:      req.WorkflowId,
		Code:            req.Code,
		Language:        req.Language,
		Args:            req.Args,
		Kind:            string(req.Kind),
		ExternalId:      req.ExternalId,
		IsTiming:        req.IsTiming,
		ScheduledTime:   req.ScheduledTime,
		Variables: slice.Map(req.Variables, func(idx int, src domain.Variables) dao.Variables {
			return dao.Variables{
				Key:    src.Key,
				Value:  src.Value,
				Secret: src.Secret,
			}
		}),
		Status:    req.Status.ToUint8(),
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Target:    req.Target,
		Handler:   req.Handler,
	}

	return t
}

func (repo *taskRepository) toDomain(req dao.Task) domain.Task {
	t := domain.Task{
		Id:            req.Id,
		ProcessInstId: req.ProcessInstId,
		CurrentNodeId: req.CurrentNodeId,
		OrderId:       req.OrderId,
		CodebookUid:   req.CodebookUid,
		CodebookName:  req.CodebookName,
		WorkflowId:    req.WorkflowId,
		Code:          req.Code,
		Args:          req.Args,
		Kind:          domain.Kind(req.Kind),
		ExternalId:    req.ExternalId,
		IsTiming:      req.IsTiming,
		ScheduledTime: req.ScheduledTime,
		Variables: slice.Map(req.Variables, func(idx int, src dao.Variables) domain.Variables {
			return domain.Variables{
				Key:    src.Key,
				Value:  src.Value,
				Secret: src.Secret,
			}
		}),
		Language:        req.Language,
		Utime:           req.Utime,
		Result:          req.Result,
		WantResult:      req.WantResult,
		TriggerPosition: req.TriggerPosition,
		Status:          domain.Status(req.Status),
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		RetryCount:      req.RetryCount,
		Target:          req.Target,
		Handler:         req.Handler,
	}

	if req.Kind == "" {
		t.Kind = domain.KAFKA
	}

	return t
}
