package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository/dao"
)

type ResourceRepository interface {
	CreateResource(ctx context.Context, req domain.Resource) (int64, error)
}

type resourceRepository struct {
	dao dao.ResourceDAO
}

func NewResourceRepository(dao dao.ResourceDAO) ResourceRepository {
	return &resourceRepository{
		dao: dao,
	}
}

func (a *resourceRepository) CreateResource(ctx context.Context, req domain.Resource) (int64, error) {
	return a.dao.CreateResource(ctx, req.Data, dao.Resource{
		ModelIdentifies: req.ModelIdentifies,
	})
}
