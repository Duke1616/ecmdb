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
	TodoList(userId, processName string, sortByAse bool, offset, limit int) (
		[]domain.Instance, error)
	CountTodo(ctx context.Context, userId, processName string) (int64, error)
	CountStartUser(ctx context.Context, userId, processName string) (int64, error)
	GetTasksByCurrentNodeId(ctx context.Context, processInstId int, currentNodeId string) ([]model.Task, error)
	ListStartUser(ctx context.Context, userId, processName string, offset, limit int) (
		[]domain.Instance, error)
	ListTaskRecord(ctx context.Context, processInstId, offset, limit int) ([]model.Task, error)
	CountTaskRecord(ctx context.Context, processInstId int) (int64, error)
	UpdateIsFinishedByPreNodeId(ctx context.Context, nodeId string, status int, comment string) error
	// ForceUpdateIsFinishedByPreNodeId 强制清理指定节点下的所有任务（包括已完成的）
	ForceUpdateIsFinishedByPreNodeId(ctx context.Context, nodeId string, status int, comment string) error
	// ForceUpdateIsFinishedByNodeId 强制清理指定节点ID的所有任务（包括已完成的）
	ForceUpdateIsFinishedByNodeId(ctx context.Context, nodeId string, status int, comment string) error
	CountReject(ctx context.Context, taskId int) (int64, error)
	ListTasksByProcInstIds(ctx context.Context, processInstIds []int, starter string) ([]domain.Instance, error)
	GetAutomationTask(ctx context.Context, currentNodeId string, processInstId int) (model.Task, error)
	GetTasksByInstUsers(ctx context.Context, processInstId int, userIds []string) ([]model.Task, error)
	GetOrderIdByVariable(ctx context.Context, processInstId int) (string, error)
	// GetProxyNodeID 获取代理转发的节点ID
	GetProxyNodeID(ctx context.Context, prevNodeID string) (model.Task, error)
	// GetProxyNodeByProcessInstId 通过流程实例ID获取 proxy 节点
	GetProxyNodeByProcessInstId(ctx context.Context, processInstId int) (model.Task, error)
	// DeleteProxyNodeByNodeId 删除指定 proxy 节点任务记录
	DeleteProxyNodeByNodeId(ctx context.Context, nodeId string) error
	// UpdateTaskPrevNodeID 修改任务的上级节点ID
	UpdateTaskPrevNodeID(ctx context.Context, taskId int, prevNodeId string) error

	// GetInstanceByID 获取流程实例详情 (用于获取版本号)
	GetInstanceByID(ctx context.Context, processInstId int) (domain.Instance, error)
	// GetProcessDefineByVersion 获取指定版本的流程定义 (包含历史版本)
	GetProcessDefineByVersion(ctx context.Context, processID, version int) (model.Process, error)
	// GetLatestProcessVersion 获取流程的最新版本号
	GetLatestProcessVersion(ctx context.Context, processID int) (int, error)
}

type processEngineRepository struct {
	engineDao dao.ProcessEngineDAO
}

func (repo *processEngineRepository) UpdateTaskPrevNodeID(ctx context.Context, taskId int, prevNodeId string) error {
	return repo.engineDao.UpdateTaskPrevNodeID(ctx, taskId, prevNodeId)
}

func (repo *processEngineRepository) GetProxyNodeID(ctx context.Context, prevNodeID string) (model.Task, error) {
	return repo.engineDao.GetProxyNodeID(ctx, prevNodeID)
}

func (repo *processEngineRepository) GetProxyNodeByProcessInstId(ctx context.Context, processInstId int) (model.Task, error) {
	return repo.engineDao.GetProxyNodeByProcessInstId(ctx, processInstId)
}

func (repo *processEngineRepository) DeleteProxyNodeByNodeId(ctx context.Context, nodeId string) error {
	return repo.engineDao.DeleteProxyNodeByNodeId(ctx, nodeId)
}

func (repo *processEngineRepository) GetInstanceByID(ctx context.Context, processInstId int) (domain.Instance, error) {
	inst, err := repo.engineDao.GetInstanceByID(ctx, processInstId)
	if err != nil {
		return domain.Instance{}, err
	}
	return repo.toDomainByInstance(inst), nil
}

func (repo *processEngineRepository) GetProcessDefineByVersion(ctx context.Context, processID, version int) (model.Process, error) {
	return repo.engineDao.GetProcessDefineByVersion(ctx, processID, version)
}

func (repo *processEngineRepository) GetLatestProcessVersion(ctx context.Context, processID int) (int, error) {
	return repo.engineDao.GetLatestProcessVersion(ctx, processID)
}

func (repo *processEngineRepository) GetTasksByCurrentNodeId(ctx context.Context, processInstId int, currentNodeId string) ([]model.Task, error) {
	return repo.engineDao.GetTasksByCurrentNodeId(ctx, processInstId, currentNodeId)
}

func (repo *processEngineRepository) GetOrderIdByVariable(ctx context.Context, processInstId int) (string, error) {
	return repo.engineDao.GetOrderIdByVariable(ctx, processInstId)
}

func (repo *processEngineRepository) GetTasksByInstUsers(ctx context.Context, processInstId int, userIds []string) ([]model.Task, error) {
	return repo.engineDao.GetTasksByInstUsers(ctx, processInstId, userIds)
}

func (repo *processEngineRepository) GetAutomationTask(ctx context.Context, currentNodeId string, processInstId int) (model.Task, error) {
	return repo.engineDao.GetAutomationTask(ctx, currentNodeId, processInstId)
}

func (repo *processEngineRepository) ListTasksByProcInstIds(ctx context.Context, processInstIds []int, starter string) (
	[]domain.Instance, error) {
	ts, err := repo.engineDao.ListTasksByProcInstId(ctx, processInstIds, starter)
	return slice.Map(ts, func(idx int, src model.Task) domain.Instance {
		return repo.toDomainByTask(src)
	}), err
}

func (repo *processEngineRepository) CountReject(ctx context.Context, taskId int) (int64, error) {
	return repo.engineDao.CountReject(ctx, taskId)
}

func (repo *processEngineRepository) UpdateIsFinishedByPreNodeId(ctx context.Context, nodeId string, status int, comment string) error {
	return repo.engineDao.UpdateIsFinishedByPreNodeId(ctx, nodeId, status, comment)
}

// ForceUpdateIsFinishedByPreNodeId 强制清理指定节点下的所有任务（包括已完成的）
func (repo *processEngineRepository) ForceUpdateIsFinishedByPreNodeId(ctx context.Context, nodeId string, status int, comment string) error {
	return repo.engineDao.ForceUpdateIsFinishedByPreNodeId(ctx, nodeId, status, comment)
}

// ForceUpdateIsFinishedByNodeId 强制清理指定节点ID的所有任务（包括已完成的）
func (repo *processEngineRepository) ForceUpdateIsFinishedByNodeId(ctx context.Context, nodeId string, status int, comment string) error {
	return repo.engineDao.ForceUpdateIsFinishedByNodeId(ctx, nodeId, status, comment)
}

func (repo *processEngineRepository) TodoList(userId, processName string, sortByAse bool, offset, limit int) ([]domain.Instance, error) {
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

func (repo *processEngineRepository) ListTaskRecord(ctx context.Context, processInstId, offset, limit int) ([]model.Task, error) {
	return repo.engineDao.ListTaskRecord(ctx, processInstId, offset, limit)
}

func (repo *processEngineRepository) CountTaskRecord(ctx context.Context, processInstId int) (int64, error) {
	return repo.engineDao.CountTaskRecord(ctx, processInstId)
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
		ApprovedBy:      req.ApprovedBy,
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
		CreateTime:      req.CreateTime,
		ApprovedBy:      req.UserID,
		Status:          req.Status,
	}
}
