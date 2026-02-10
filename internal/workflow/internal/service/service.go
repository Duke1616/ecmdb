package service

import (
	"context"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/domain"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/repository"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	// Create 创建流程定义
	Create(ctx context.Context, req domain.Workflow) (int64, error)
	// List 分页查询流程定义列表
	List(ctx context.Context, offset, limit int64) ([]domain.Workflow, int64, error)
	// Find 根据ID查询流程定义（返回最新版元数据）
	Find(ctx context.Context, id int64) (domain.Workflow, error)
	// Update 更新流程定义
	Update(ctx context.Context, req domain.Workflow) (int64, error)
	// Delete 删除流程定义
	Delete(ctx context.Context, id int64) (int64, error)
	// Deploy 发布流程到引擎并创建快照
	Deploy(ctx context.Context, flow domain.Workflow) error
	// FindByKeyword 根据关键字搜索流程
	FindByKeyword(ctx context.Context, keyword string, offset, limit int64) ([]domain.Workflow, int64, error)
	// FindPassEdgeIds 查找所有已经完成的边id
	FindPassEdgeIds(ctx context.Context, wf domain.Workflow, tasks []model.Task) ([]string, error)
	// GetAutomationProperty 获取指定节点的自动化属性配置
	GetAutomationProperty(workflow easyflow.Workflow, nodeId string) (easyflow.AutomationProperty, error)
	// GetWorkflowSnapshot 获取指定版本的流程快照（原子语义，直接返回快照数据）
	GetWorkflowSnapshot(ctx context.Context, processID, version int) (domain.Workflow, error)
	// FindInstanceFlow 获取流程实例对应的流程定义（包含历史快照回溯与降级逻辑）
	FindInstanceFlow(ctx context.Context, workflowID int64, processID, version int) (domain.Workflow, error)

	// AdminNotifyBinding 管理侧：通知绑定配置
	AdminNotifyBinding() AdminNotifyBindingService
}

type AdminNotifyBindingService interface {
	// Create 创建通知绑定
	Create(ctx context.Context, n domain.NotifyBinding) (int64, error)
	// Update 更新通知绑定
	Update(ctx context.Context, n domain.NotifyBinding) (int64, error)
	// Delete 删除通知绑定
	Delete(ctx context.Context, id int64) (int64, error)
	// List 查询流程下的绑定配置
	List(ctx context.Context, workflowId int64) ([]domain.NotifyBinding, error)
	// GetEffective 获取最终生效的配置 (含默认兜底逻辑)
	GetEffective(ctx context.Context, workflowId int64, notifyType domain.NotifyType, channel string) (domain.NotifyBinding, error)
}

type service struct {
	repo         repository.WorkflowRepository
	bindingRepo  repository.NotifyBindingRepository
	engineSvc    engine.Service
	engineCovert easyflow.ProcessEngineConvert
}

// 内部实现：Binding Service
type adminNotifyBindingService struct {
	repo repository.NotifyBindingRepository
}

func (s *adminNotifyBindingService) Create(ctx context.Context, n domain.NotifyBinding) (int64, error) {
	return s.repo.Create(ctx, n)
}

func (s *adminNotifyBindingService) Update(ctx context.Context, n domain.NotifyBinding) (int64, error) {
	return s.repo.Update(ctx, n)
}

func (s *adminNotifyBindingService) Delete(ctx context.Context, id int64) (int64, error) {
	return s.repo.Delete(ctx, id)
}

func (s *adminNotifyBindingService) List(ctx context.Context, workflowId int64) ([]domain.NotifyBinding, error) {
	return s.repo.List(ctx, workflowId)
}

func (s *adminNotifyBindingService) GetEffective(ctx context.Context, workflowId int64, notifyType domain.NotifyType, channel string) (domain.NotifyBinding, error) {
	return s.repo.GetEffective(ctx, workflowId, notifyType, channel)
}

func (s *service) GetAutomationProperty(workflow easyflow.Workflow, nodeId string) (easyflow.AutomationProperty, error) {
	return s.engineCovert.GetAutomationProperty(workflow, nodeId)
}

func NewService(repo repository.WorkflowRepository, bindingRepo repository.NotifyBindingRepository,
	engineSvc engine.Service, engineCovert easyflow.ProcessEngineConvert) Service {
	return &service{
		repo:         repo,
		bindingRepo:  bindingRepo,
		engineSvc:    engineSvc,
		engineCovert: engineCovert,
	}
}

func (s *service) AdminNotifyBinding() AdminNotifyBindingService {
	return &adminNotifyBindingService{
		repo: s.bindingRepo,
	}
}

func (s *service) Create(ctx context.Context, req domain.Workflow) (int64, error) {
	return s.repo.Create(ctx, req)
}

func (s *service) List(ctx context.Context, offset, limit int64) ([]domain.Workflow, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Workflow
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.List(ctx, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) Find(ctx context.Context, id int64) (domain.Workflow, error) {
	return s.repo.Find(ctx, id)
}

func (s *service) Update(ctx context.Context, req domain.Workflow) (int64, error) {
	return s.repo.Update(ctx, req)
}

func (s *service) Delete(ctx context.Context, id int64) (int64, error) {
	return s.repo.Delete(ctx, id)
}

func (s *service) FindPassEdgeIds(ctx context.Context, wf domain.Workflow, tasks []model.Task) ([]string, error) {
	return s.engineCovert.Edge(s.toEasyWorkflow(wf), tasks)
}

func (s *service) Deploy(ctx context.Context, wf domain.Workflow) error {
	// 发布到流程引擎
	processId, err := s.engineCovert.Deploy(s.toEasyWorkflow(wf))
	if err != nil {
		return err
	}

	// 绑定此流程对应引擎的ID, 为了后续查询数据详情使用
	if err = s.repo.UpdateProcessId(ctx, wf.Id, processId); err != nil {
		return err
	}

	// Double Check: 获取刚刚创建的流程版本号
	version, err := s.engineSvc.GetLatestProcessVersion(ctx, processId)
	if err != nil {
		return err
	}

	// 创建快照 (记录此刻的 FlowData 与 ProcessID/Version 的关系)
	return s.repo.CreateSnapshot(ctx, wf, processId, version)
}

func (s *service) GetWorkflowSnapshot(ctx context.Context, processID, version int) (domain.Workflow, error) {
	return s.repo.FindSnapshot(ctx, processID, version)
}

func (s *service) FindInstanceFlow(ctx context.Context, workflowID int64, processID, version int) (domain.Workflow, error) {
	// 1. 获取最新版元数据
	latest, err := s.Find(ctx, workflowID)
	// 如果主记录找不到了，尝试纯靠快照恢复
	if err != nil {
		snapshot, snapErr := s.repo.FindSnapshot(ctx, processID, version)
		if snapErr != nil {
			return domain.Workflow{}, err
		}
		// 此时 snapshot 已经是 domain.Workflow 类型
		return snapshot, nil
	}

	// 2. 尝试读取锁定的快照
	snapshot, err := s.repo.FindSnapshot(ctx, processID, version)
	if err == nil {
		// 覆盖 FlowData (Snapshot 现在是 domain.Workflow)
		latest.FlowData = snapshot.FlowData
	}

	return latest, nil
}

func (s *service) FindByKeyword(ctx context.Context, keyword string, offset, limit int64) ([]domain.Workflow, int64, error) {
	var (
		eg    errgroup.Group
		ws    []domain.Workflow
		total int64
	)
	eg.Go(func() error {
		var err error
		ws, err = s.repo.FindByKeyword(ctx, keyword, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountByKeyword(ctx, keyword)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ws, total, err
	}
	return ws, total, nil
}

func (s *service) toEasyWorkflow(wf domain.Workflow) easyflow.Workflow {
	return easyflow.Workflow{
		Id:    wf.Id,
		Name:  wf.Name,
		Owner: wf.Owner,
		FlowData: easyflow.LogicFlow{
			Edges: wf.FlowData.Edges,
			Nodes: wf.FlowData.Nodes,
		},
	}
}
