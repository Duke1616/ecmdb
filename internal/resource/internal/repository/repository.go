package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository/dao"
	"github.com/Duke1616/ecmdb/pkg/mongox"
)

type ResourceRepository interface {
	CreateResource(ctx context.Context, req domain.Resource) (int64, error)
	FindResourceById(ctx context.Context, dmAttr domain.DetailResource) ([]mongox.MapStr, error)
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
	return r.dao.CreateResource(ctx, dao.Resource{
		ModelUID: req.ModelUID,
		Name:     req.Name,
		Data:     req.Data,
	})
}

func (r *resourceRepository) FindResourceById(ctx context.Context, dmAttr domain.DetailResource) ([]mongox.MapStr, error) {
	resource, err := r.dao.FindResourceById(ctx, dmAttr)
	if err != nil {
		return nil, err
	}

	return resource, nil
}
