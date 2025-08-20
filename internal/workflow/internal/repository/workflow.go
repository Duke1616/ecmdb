package repository

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/workflow/internal/domain"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type WorkflowRepository interface {
	Create(ctx context.Context, req domain.Workflow) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]domain.Workflow, error)
	Total(ctx context.Context) (int64, error)
	Update(ctx context.Context, req domain.Workflow) (int64, error)
	UpdateProcessId(ctx context.Context, id int64, processId int) error
	Delete(ctx context.Context, id int64) (int64, error)
	Find(ctx context.Context, id int64) (domain.Workflow, error)
}

func NewWorkflowRepository(dao dao.WorkflowDAO) WorkflowRepository {
	return &workflowRepository{
		dao: dao,
	}
}

type workflowRepository struct {
	dao dao.WorkflowDAO
}

func (repo *workflowRepository) Create(ctx context.Context, req domain.Workflow) (int64, error) {
	return repo.dao.Create(ctx, repo.toEntity(req))
}

func (repo *workflowRepository) List(ctx context.Context, offset, limit int64) ([]domain.Workflow, error) {
	ws, err := repo.dao.List(ctx, offset, limit)
	return slice.Map(ws, func(idx int, src dao.Workflow) domain.Workflow {
		return repo.toDomain(src)
	}), err
}
func (repo *workflowRepository) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func (repo *workflowRepository) Find(ctx context.Context, id int64) (domain.Workflow, error) {
	w, err := repo.dao.Find(ctx, id)
	return repo.toDomain(w), err
}

func (repo *workflowRepository) Update(ctx context.Context, req domain.Workflow) (int64, error) {
	return repo.dao.Update(ctx, repo.toEntity(req))
}

func (repo *workflowRepository) Delete(ctx context.Context, id int64) (int64, error) {
	return repo.dao.Delete(ctx, id)
}

func (repo *workflowRepository) UpdateProcessId(ctx context.Context, id int64, processId int) error {
	return repo.dao.UpdateProcessId(ctx, id, processId)
}

func (repo *workflowRepository) toEntity(req domain.Workflow) dao.Workflow {
	return dao.Workflow{
		Id:           req.Id,
		TemplateId:   req.TemplateId,
		Name:         req.Name,
		Icon:         req.Icon,
		Owner:        req.Owner,
		Desc:         req.Desc,
		NotifyMethod: req.NotifyMethod.ToUint8(),
		IsNotify:     req.IsNotify,
		FlowData: dao.LogicFlow{
			Edges: req.FlowData.Edges,
			Nodes: req.FlowData.Nodes,
		},
	}
}

func (repo *workflowRepository) toDomain(req dao.Workflow) domain.Workflow {
	return domain.Workflow{
		Id:           req.Id,
		TemplateId:   req.TemplateId,
		Name:         req.Name,
		Icon:         req.Icon,
		Owner:        req.Owner,
		Desc:         req.Desc,
		ProcessId:    req.ProcessId,
		NotifyMethod: domain.NotifyMethod(req.NotifyMethod),
		IsNotify:     req.IsNotify,
		FlowData: domain.LogicFlow{
			Edges: req.FlowData.Edges,
			Nodes: req.FlowData.Nodes,
		},
	}
}
