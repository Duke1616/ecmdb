package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/discovery/internal/domain"
	"github.com/Duke1616/ecmdb/internal/discovery/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type DiscoveryRepository interface {
	Create(ctx context.Context, req domain.Discovery) (int64, error)
	Update(ctx context.Context, req domain.Discovery) (int64, error)
	Delete(ctx context.Context, id int64) (int64, error)
	ListByTemplateId(ctx context.Context, offset, limit int64, templateId int64) ([]domain.Discovery, error)
	CountByTemplateId(ctx context.Context, templateId int64) (int64, error)
}

type discoveryRepository struct {
	dao dao.DiscoveryDAO
}

func (repo *discoveryRepository) Delete(ctx context.Context, id int64) (int64, error) {
	return repo.dao.Delete(ctx, id)
}

func (repo *discoveryRepository) Create(ctx context.Context, req domain.Discovery) (int64, error) {
	return repo.dao.Update(ctx, repo.toEntity(req))
}

func (repo *discoveryRepository) Update(ctx context.Context, req domain.Discovery) (int64, error) {
	return repo.dao.Update(ctx, repo.toEntity(req))
}

func (repo *discoveryRepository) ListByTemplateId(ctx context.Context, offset, limit int64, templateId int64) ([]domain.Discovery, error) {
	ds, err := repo.dao.ListByTemplateId(ctx, offset, limit, templateId)
	return slice.Map(ds, func(idx int, src dao.Discovery) domain.Discovery {
		return repo.toDomain(src)
	}), err
}

func (repo *discoveryRepository) CountByTemplateId(ctx context.Context, templateId int64) (int64, error) {
	return repo.dao.CountByTemplateId(ctx, templateId)
}

func NewDiscoveryRepository(dao dao.DiscoveryDAO) DiscoveryRepository {
	return &discoveryRepository{
		dao: dao,
	}
}

func (repo *discoveryRepository) toDomain(src dao.Discovery) domain.Discovery {
	return domain.Discovery{
		Id:         src.Id,
		TemplateId: src.TemplateId,
		RunnerId:   src.RunnerId,
		RunnerName: src.RunnerName,
		Title:      src.Title,
		Field:      src.Field,
		Value:      src.Value,
	}
}

func (repo *discoveryRepository) toEntity(src domain.Discovery) dao.Discovery {
	return dao.Discovery{
		Id:         src.Id,
		TemplateId: src.TemplateId,
		RunnerId:   src.RunnerId,
		RunnerName: src.RunnerName,
		Title:      src.Title,
		Field:      src.Field,
		Value:      src.Value,
	}
}
