package service

import (
	"context"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/domain"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/repository"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	Create(ctx context.Context, req domain.Workflow) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]domain.Workflow, int64, error)
	Find(ctx context.Context, id int64) (domain.Workflow, error)
	Update(ctx context.Context, req domain.Workflow) (int64, error)
	Delete(ctx context.Context, id int64) (int64, error)
	Deploy(ctx context.Context, flow domain.Workflow) error

	// FindPassEdgeIds 查找所有已经完成的边id
	FindPassEdgeIds(ctx context.Context, wf domain.Workflow, tasks []model.Task) ([]string, error)
}

type service struct {
	repo         repository.WorkflowRepository
	engineCovert easyflow.ProcessEngineConvert
}

func NewService(repo repository.WorkflowRepository, engineCovert easyflow.ProcessEngineConvert) Service {
	return &service{
		repo:         repo,
		engineCovert: engineCovert,
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
	return s.repo.UpdateProcessId(ctx, wf.Id, processId)
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
