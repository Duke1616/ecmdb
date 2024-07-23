package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/runner"
	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
)

type Service interface {
	// CreateTask 创建自动化任务
	CreateTask(ctx context.Context, processInstId int, nodeId string) error
}

type service struct {
	repo        repository.TaskRepository
	orderSvc    order.Service
	workflowSvc workflow.Service
	codebookSvc codebook.Service
	runnerSvc   runner.Service
	workerSvc   worker.Service
}

func NewService(repo repository.TaskRepository, orderSvc order.Service, workflowSvc workflow.Service,
	codebookSvc codebook.Service, runnerSvc runner.Service, workerSvc worker.Service) Service {
	return &service{
		repo:        repo,
		orderSvc:    orderSvc,
		workflowSvc: workflowSvc,
		codebookSvc: codebookSvc,
		runnerSvc:   runnerSvc,
		workerSvc:   workerSvc,
	}
}

func (s *service) CreateTask(ctx context.Context, processInstId int, nodeId string) error {
	orderResp, err := s.orderSvc.DetailByProcessInstId(ctx, processInstId)
	if err != nil {
		return err
	}

	// TODO 后期引用[实时] OR [定时]执行逻辑  目前都是立即执行
	flow, err := s.workflowSvc.Find(ctx, orderResp.WorkflowId)
	if err != nil {
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
	}, nodeId)

	if err != nil {
		return err
	}

	// 查询调度节点
	runnerResp, err := s.runnerSvc.FindByCodebookUid(ctx, automation.CodebookUid, automation.Tag)
	if err != nil {
		return err
	}

	// 查询执行代码
	codebookResp, err := s.codebookSvc.FindByUid(ctx, runnerResp.CodebookUid)
	if err != nil {
		return err
	}

	// 查看工作节点Topic
	workerResp, err := s.workerSvc.FindByName(ctx, runnerResp.WorkerName)
	if err != nil {
		return err
	}

	// TODO 查看节点状态，禁用 离线 是否可以堆积到消息队列中
	switch workerResp.Status {
	case worker.STOPPING:
	case worker.OFFLINE:

	}

	// TODO 创建一份任务到数据库中，后续执行失败，重试机制
	_, err = s.repo.CreateTask(ctx, domain.Task{
		// 字段可有可无
		ProcessInstId: processInstId,
		WorkerName:    workerResp.Name,
		WorkflowId:    flow.Id,
		CodebookUid:   codebookResp.Identifier,

		// 必传字段
		OrderId:  orderResp.Id,
		Code:     codebookResp.Code,
		Topic:    workerResp.Topic,
		Language: codebookResp.Language,
	})
	if err != nil {
		return err
	}

	// 发送任务
	return s.workerSvc.Execute(ctx, worker.Execute{
		Name:     "自动化流程",
		Topic:    workerResp.Topic,
		Code:     codebookResp.Code,
		Language: codebookResp.Language,
	})
}
