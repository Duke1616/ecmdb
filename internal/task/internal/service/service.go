package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/runner"
	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/gotomicro/ego/core/elog"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	// CreateTask 创建任务
	CreateTask(ctx context.Context, processInstId int, nodeId string) error
	// StartTask 启动任务
	StartTask(ctx context.Context, processInstId int, nodeId string) error
	RetryTask(ctx context.Context, id int64) error
	UpdateTaskStatus(ctx context.Context, req domain.TaskResult) (int64, error)
	UpdateArgs(ctx context.Context, id int64, args map[string]interface{}) (int64, error)
	UpdateVariables(ctx context.Context, id int64, variables string) (int64, error)
	// ListTaskByStatus 列表任务
	ListTaskByStatus(ctx context.Context, offset, limit int64, status uint8) ([]domain.Task, int64, error)
	ListTask(ctx context.Context, offset, limit int64) ([]domain.Task, int64, error)

	ListTasksByCtime(ctx context.Context, offset, limit int64, ctime int64) ([]domain.Task, int64, error)
}

type service struct {
	repo        repository.TaskRepository
	logger      *elog.Component
	orderSvc    order.Service
	engineSvc   engine.Service
	workflowSvc workflow.Service
	codebookSvc codebook.Service
	runnerSvc   runner.Service
	workerSvc   worker.Service
}

func (s *service) ListTasksByCtime(ctx context.Context, offset, limit int64, ctime int64) ([]domain.Task, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Task
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListTasksByCtime(ctx, offset, limit, ctime)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalByCtime(ctx, ctime)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) ListTask(ctx context.Context, offset, limit int64) ([]domain.Task, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Task
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListTask(ctx, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx, 0)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) CreateTask(ctx context.Context, processInstId int, nodeId string) error {
	_, err := s.repo.CreateTask(ctx, domain.Task{
		ProcessInstId:   processInstId,
		TriggerPosition: "任务等待",
		CurrentNodeId:   nodeId,
		Status:          domain.WAITING,
	})

	return err
}

func NewService(repo repository.TaskRepository, orderSvc order.Service, workflowSvc workflow.Service,
	codebookSvc codebook.Service, runnerSvc runner.Service, workerSvc worker.Service, engineSvc engine.Service) Service {
	return &service{
		repo:        repo,
		logger:      elog.DefaultLogger,
		orderSvc:    orderSvc,
		workflowSvc: workflowSvc,
		codebookSvc: codebookSvc,
		runnerSvc:   runnerSvc,
		workerSvc:   workerSvc,
		engineSvc:   engineSvc,
	}
}

func (s *service) StartTask(ctx context.Context, processInstId int, nodeId string) error {
	// 先创建任务、以防后续失败，导致无法溯源
	task, err := s.repo.FindByProcessInstId(ctx, processInstId, nodeId)
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return s.process(ctx, task)
	}

	// 新建工作记录
	taskId, err := s.repo.CreateTask(ctx, domain.Task{
		ProcessInstId:   processInstId,
		TriggerPosition: "开始节点",
		CurrentNodeId:   nodeId,
		Status:          domain.SCHEDULE,
	})

	// 追加赋值
	task.Id = taskId
	task.ProcessInstId = processInstId
	task.CurrentNodeId = nodeId

	if err != nil {
		elog.Error("创建任务失败",
			elog.Any("错误信息", err),
			elog.Any("流程实例ID", processInstId),
			elog.Any("当前节点ID", nodeId),
		)
		return err
	}

	return s.process(ctx, task)
}

func (s *service) RetryTask(ctx context.Context, id int64) error {
	task, err := s.repo.FindById(ctx, id)

	if err != nil {
		elog.Error("重试任务失败",
			elog.Any("错误信息", err),
			elog.Any("任务ID", id),
		)
		return err
	}

	// 证明这个任务是失败的，应该重新获取数据
	if task.OrderId == 0 {
		return s.process(ctx, task)
	}

	return s.retry(ctx, task)
}

func (s *service) UpdateArgs(ctx context.Context, id int64, args map[string]interface{}) (int64, error) {
	return s.repo.UpdateArgs(ctx, id, args)
}

func (s *service) UpdateVariables(ctx context.Context, id int64, variables string) (int64, error) {
	return s.repo.UpdateVariables(ctx, id, variables)
}

func (s *service) ListTaskByStatus(ctx context.Context, offset, limit int64, status uint8) ([]domain.Task, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Task
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListTaskByStatus(ctx, offset, limit, status)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx, status)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) UpdateTaskStatus(ctx context.Context, req domain.TaskResult) (int64, error) {
	return s.repo.UpdateTaskStatus(ctx, req)
}

func (s *service) retry(ctx context.Context, task domain.Task) error {
	return s.workerSvc.Execute(ctx, worker.Execute{
		TaskId:    task.Id,
		Topic:     task.Topic,
		Code:      task.Code,
		Language:  task.Language,
		Args:      task.Args,
		Variables: task.Variables,
	})
}

func (s *service) process(ctx context.Context, task domain.Task) error {
	// 获取工单信息
	orderResp, err := s.orderSvc.DetailByProcessInstId(ctx, task.ProcessInstId)
	if err != nil {
		_, _ = s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			TriggerPosition: "获取工单失败",
			Status:          domain.FAILED,
			Result:          err.Error(),
		})
		return err
	}

	// TODO 后期引用[实时] OR [定时]执行逻辑  目前都是立即执行
	flow, err := s.workflowSvc.Find(ctx, orderResp.WorkflowId)
	if err != nil {
		_, _ = s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			TriggerPosition: "获取流程信息失败",
			Status:          domain.FAILED,
			Result:          err.Error(),
		})
		return err
	}

	// 获取自动化提交信息
	automation, err := s.workflowSvc.GetAutomationProperty(easyflow.Workflow{
		Id:    flow.Id,
		Name:  flow.Name,
		Owner: flow.Owner,
		FlowData: easyflow.LogicFlow{
			Edges: flow.FlowData.Edges,
			Nodes: flow.FlowData.Nodes,
		},
	}, task.CurrentNodeId)

	if err != nil {
		_, _ = s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			TriggerPosition: "提取自动化信息失败",
			Status:          domain.FAILED,
			Result:          err.Error(),
		})
		return err
	}

	// 查看调度工作节点
	runnerResp, err := s.runnerSvc.FindByCodebookUid(ctx, automation.CodebookUid, automation.Tag)
	if err != nil {
		_, _ = s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			TriggerPosition: "获取调度节点失败",
			Status:          domain.PENDING,
			Result:          err.Error(),
		})
		return err
	}

	// 查看工作节点Topic
	workerResp, err := s.workerSvc.FindByName(ctx, runnerResp.WorkerName)
	if err != nil {
		_, _ = s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			TriggerPosition: "获取工作节点失败",
			Status:          domain.FAILED,
			Result:          err.Error(),
		})
		return err
	}

	// 查询执行代码
	codebookResp, err := s.codebookSvc.FindByUid(ctx, runnerResp.CodebookUid)
	if err != nil {
		_, _ = s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			TriggerPosition: "获取任务模版失败",
			Status:          domain.FAILED,
			Result:          err.Error(),
		})
		return err
	}

	// TODO 查看节点状态，禁用 离线 是否可以堆积到消息队列中
	switch workerResp.Status {
	case worker.STOPPING:
	case worker.OFFLINE:

	}

	// TODO 创建一份任务到数据库中，后续执行失败，重试机制
	vars, _ := json.Marshal(runnerResp.Variables)
	_, err = s.repo.UpdateTask(ctx, domain.Task{
		// 字段可有可无
		Id:            task.Id,
		ProcessInstId: task.ProcessInstId,
		WorkerName:    workerResp.Name,
		WorkflowId:    flow.Id,
		CodebookUid:   codebookResp.Identifier,

		// 必传字段
		OrderId:   orderResp.Id,
		Code:      codebookResp.Code,
		Topic:     workerResp.Topic,
		Language:  codebookResp.Language,
		Status:    domain.RUNNING,
		Args:      orderResp.Data,
		Variables: string(vars),
	})
	if err != nil {
		return err
	}

	// 发送任务执行
	return s.workerSvc.Execute(ctx, worker.Execute{
		TaskId:    task.Id,
		Topic:     workerResp.Topic,
		Code:      codebookResp.Code,
		Language:  codebookResp.Language,
		Args:      orderResp.Data,
		Variables: string(vars),
	})
}
