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

func (repo *workflowRepository) toEntity(req domain.Workflow) dao.Workflow {
	return dao.Workflow{
		TemplateId: req.TemplateId,
		Name:       req.Name,
		Icon:       req.Icon,
		Owner:      req.Owner,
		Desc:       req.Desc,
		FlowData:   req.FlowData,
	}
}

func (repo *workflowRepository) toDomain(req dao.Workflow) domain.Workflow {
	return domain.Workflow{
		Id:         req.Id,
		TemplateId: req.TemplateId,
		Name:       req.Name,
		Icon:       req.Icon,
		Owner:      req.Owner,
		Desc:       req.Desc,
		FlowData:   req.FlowData,
	}
}
