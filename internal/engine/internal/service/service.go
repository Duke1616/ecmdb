package service

import (
	"context"

	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/engine/internal/domain"
	"github.com/Duke1616/ecmdb/internal/engine/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	// ListTodoTasks 查看todo任务
	ListTodoTasks(ctx context.Context, userId, processName string, sortByAse bool, offset, limit int) (
		[]domain.Instance, int64, error)

	ListByStartUser(ctx context.Context, userId, processName string, offset, limit int) (
		[]domain.Instance, int64, error)
	// TaskRecord 工单任务变更记录
	TaskRecord(ctx context.Context, processInstId, offset, limit int) ([]model.Task, int64, error)
	IsReject(ctx context.Context, taskId int) (bool, error)
	// GetTasksByCurrentNodeId 获取当前节点下的所有任务
	GetTasksByCurrentNodeId(ctx context.Context, processInstId int, currentNodeId string) ([]model.Task, error)
	// UpdateIsFinishedByPreNodeId 系统修改 finished 状态
	UpdateIsFinishedByPreNodeId(ctx context.Context, nodeId string, status int, comment string) error
	// Pass 通过
	Pass(ctx context.Context, taskId int, comment string) error
	// Reject 驳回
	Reject(ctx context.Context, taskId int, comment string) error
	// ListPendingStepsOfMyTask 列出我的任务待处理步骤
	ListPendingStepsOfMyTask(ctx context.Context, processInstIds []int, starter string) ([]domain.Instance, error)
	// GetAutomationTask 获取自动化完成任务
	GetAutomationTask(ctx context.Context, currentNodeId string, processInstId int) (model.Task, error)
	// GetTasksByInstUsers 获取指定流程 + 用户的任务
	GetTasksByInstUsers(ctx context.Context, processInstId int, userIds []string) ([]model.Task, error)
	// GetOrderIdByVariable 获取工单ID，进行流程绑定
	GetOrderIdByVariable(ctx context.Context, processInstId int) (string, error)
	// Revoke 撤销工单
	Revoke(ctx context.Context, instanceId int, userId string, force bool) error
	// Upstream 获取所有上游节点
	Upstream(ctx context.Context, taskId int) ([]model.Node, error)
	// TaskInfo 获取任务详情
	TaskInfo(ctx context.Context, taskId int) (model.Task, error)
	// GetProxyNodeID 获取代理转发的节点ID
	GetProxyNodeID(ctx context.Context, prevNodeID string) (string, error)
	// UpdateTaskPrevNodeID 修改任务节点ID
	UpdateTaskPrevNodeID(ctx context.Context, taskId int, prevNodeId string) error
}

type service struct {
	repo repository.ProcessEngineRepository
}

func (s *service) GetProxyNodeID(ctx context.Context, prevNodeID string) (string, error) {
	procTask, err := s.repo.GetProxyNodeID(ctx, prevNodeID)
	return procTask.PrevNodeID, err
}

func (s *service) TaskInfo(ctx context.Context, taskId int) (model.Task, error) {
	return engine.GetTaskInfo(taskId)
}

func (s *service) GetTasksByCurrentNodeId(ctx context.Context, processInstId int, currentNodeId string) ([]model.Task, error) {
	return s.repo.GetTasksByCurrentNodeId(ctx, processInstId, currentNodeId)
}

func (s *service) Upstream(ctx context.Context, taskId int) ([]model.Node, error) {
	return engine.TaskUpstreamNodeList(taskId)
}

func (s *service) Revoke(ctx context.Context, instanceId int, userId string, force bool) error {
	return engine.InstanceRevoke(instanceId, force, userId)
}

func (s *service) GetOrderIdByVariable(ctx context.Context, processInstId int) (string, error) {
	return s.repo.GetOrderIdByVariable(ctx, processInstId)
}

func (s *service) GetTasksByInstUsers(ctx context.Context, processInstId int, userIds []string) ([]model.Task, error) {
	return s.repo.GetTasksByInstUsers(ctx, processInstId, userIds)
}

func (s *service) GetAutomationTask(ctx context.Context, currentNodeId string, processInstId int) (model.Task, error) {
	return s.repo.GetAutomationTask(ctx, currentNodeId, processInstId)
}

func (s *service) ListPendingStepsOfMyTask(ctx context.Context, processInstIds []int, starter string) (
	[]domain.Instance, error) {
	return s.repo.ListTasksByProcInstIds(ctx, processInstIds, starter)
}

func (s *service) IsReject(ctx context.Context, taskId int) (bool, error) {
	total, err := s.repo.CountReject(ctx, taskId)

	if total >= 1 {
		return true, err
	}

	return false, err
}

func (s *service) UpdateIsFinishedByPreNodeId(ctx context.Context, nodeId string, status int, comment string) error {
	return s.repo.UpdateIsFinishedByPreNodeId(ctx, nodeId, status, comment)
}

func (s *service) Reject(ctx context.Context, taskId int, comment string) error {
	return engine.TaskReject(taskId, comment, "")
}

func (s *service) Pass(ctx context.Context, taskId int, comment string) error {
	return engine.TaskPass(taskId, comment, "", false)
}

func (s *service) UpdateTaskPrevNodeID(ctx context.Context, taskId int, prevNodeId string) error {
	return s.repo.UpdateTaskPrevNodeID(ctx, taskId, prevNodeId)
}

func (s *service) TaskRecord(ctx context.Context, processInstId, offset, limit int) ([]model.Task, int64, error) {
	var (
		eg      errgroup.Group
		records []model.Task
		total   int64
	)
	eg.Go(func() error {
		var err error
		records, err = s.repo.ListTaskRecord(ctx, processInstId, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountTaskRecord(ctx, processInstId)
		return err
	})
	if err := eg.Wait(); err != nil {
		return records, total, err
	}
	return records, total, nil
}

func NewService(repo repository.ProcessEngineRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) ListTodoTasks(ctx context.Context, userId, processName string, sortByAse bool, offset, limit int) (
	[]domain.Instance, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Instance
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.TodoList(userId, processName, sortByAse, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountTodo(ctx, userId, processName)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) ListByStartUser(ctx context.Context, userId, processName string, offset,
	limit int) ([]domain.Instance, int64, error) {

	var (
		eg    errgroup.Group
		ts    []domain.Instance
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListStartUser(ctx, userId, processName, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountStartUser(ctx, userId, processName)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}
