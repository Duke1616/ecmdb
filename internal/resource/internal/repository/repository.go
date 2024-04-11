package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository/dao"
)

type ResourceRepository interface {
	CreateResource(ctx context.Context, req domain.Resource) (int64, error)
	FindResourceById(ctx context.Context, id int64, modelIdentifies string) (domain.Resource, error)
}

type resourceRepository struct {
	dao dao.ResourceDAO
}

func NewResourceRepository(dao dao.ResourceDAO) ResourceRepository {
	return &resourceRepository{
		dao: dao,
	}
}

func (r *resourceRepository) CreateResource(ctx context.Context, req domain.Resource) (int64, error) {
	return r.dao.CreateResource(ctx, req.Data, dao.Resource{
		ModelIdentifies: req.ModelIdentifies,
	})
}

func (r *resourceRepository) FindResourceById(ctx context.Context, id int64, modelIdentifies string) (domain.Resource, error) {
	resource, err := r.dao.FindResourceById(ctx, id, modelIdentifies)
	if err != nil {
		return domain.Resource{}, err
	}

	return domain.Resource{
		ID:              resource.Id,
		ModelIdentifies: resource.ModelIdentifies,
		Data:            resource.Data,
	}, nil
}
