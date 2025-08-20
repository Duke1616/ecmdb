package repository

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
	"github.com/Duke1616/ecmdb/internal/template/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type TemplateGroupRepository interface {
	Create(ctx context.Context, req domain.TemplateGroup) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]domain.TemplateGroup, error)
	Total(ctx context.Context) (int64, error)
	ListByIds(ctx context.Context, ids []int64) ([]domain.TemplateGroup, error)
}

func NewTemplateGroupRepository(dao dao.TemplateGroupDAO) TemplateGroupRepository {
	return &templateGroupRepository{
		dao: dao,
	}
}

type templateGroupRepository struct {
	dao dao.TemplateGroupDAO
}

func (repo *templateGroupRepository) Create(ctx context.Context, req domain.TemplateGroup) (int64, error) {
	return repo.dao.Create(ctx, repo.toEntity(req))
}

func (repo *templateGroupRepository) List(ctx context.Context, offset, limit int64) ([]domain.TemplateGroup, error) {
	ts, err := repo.dao.List(ctx, offset, limit)
	return slice.Map(ts, func(idx int, src dao.TemplateGroup) domain.TemplateGroup {
		return repo.toDomain(src)
	}), err
}

func (repo *templateGroupRepository) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func (repo *templateGroupRepository) ListByIds(ctx context.Context, ids []int64) ([]domain.TemplateGroup, error) {
	ts, err := repo.dao.ListByIds(ctx, ids)
	return slice.Map(ts, func(idx int, src dao.TemplateGroup) domain.TemplateGroup {
		return repo.toDomain(src)
	}), err
}

func (repo *templateGroupRepository) toEntity(req domain.TemplateGroup) dao.TemplateGroup {
	return dao.TemplateGroup{
		Name: req.Name,
		Icon: req.Icon,
	}
}

func (repo *templateGroupRepository) toDomain(req dao.TemplateGroup) domain.TemplateGroup {
	return domain.TemplateGroup{
		Id:   req.Id,
		Name: req.Name,
		Icon: req.Icon,
	}
}
