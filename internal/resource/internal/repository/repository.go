package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository/dao"
	"github.com/Duke1616/ecmdb/pkg/mongox"
)

type ResourceRepository interface {
	CreateResource(ctx context.Context, req domain.Resource) (int64, error)
	FindResourceById(ctx context.Context, projection map[string]int, id int64) ([]mongox.MapStr, error)

	ListResourcesByIds(ctx context.Context, projection map[string]int, ids []int64) ([]domain.Resource, error)
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

func (r *resourceRepository) FindResourceById(ctx context.Context, projection map[string]int, id int64) ([]mongox.MapStr, error) {
	return r.dao.FindResourceById(ctx, projection, id)
}

func (r *resourceRepository) ListResourcesByIds(ctx context.Context, projection map[string]int, ids []int64) ([]domain.Resource, error) {
	resources, err := r.dao.ListResourcesByIds(ctx, projection, ids)
	if err != nil {
		return nil, err
	}

	res := make([]domain.Resource, 0, len(resources))

	for _, val := range resources {
		res = append(res, domain.Resource{
			ID:       val.ID,
			ModelUID: val.ModelUID,
			Data:     val.Data,
			Name:     val.Name,
		})
	}

	return res, nil
}
