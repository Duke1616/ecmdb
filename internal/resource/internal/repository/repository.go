package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository/dao"
)

type ResourceRepository interface {
	CreateResource(ctx context.Context, req domain.Resource) (int64, error)
	FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error)
	ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]domain.Resource, error)

	ListResourcesByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error)

	ListExcludeResource(ctx context.Context, fields []string, modelUid string, offset, limit int64, ids []int64) ([]domain.Resource, error)
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
	return r.dao.CreateResource(ctx, r.toEntity(req))
}

func (r *resourceRepository) FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error) {
	rs, err := r.dao.FindResourceById(ctx, fields, id)
	if err != nil {
		return domain.Resource{}, err
	}

	return domain.Resource{
		ID:       rs.ID,
		Name:     rs.Name,
		ModelUID: rs.ModelUID,
		Data:     rs.Data,
	}, nil
}

func (r *resourceRepository) ListResourcesByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error) {
	resources, err := r.dao.ListResourcesByIds(ctx, fields, ids)
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

func (r *resourceRepository) ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]domain.Resource, error) {
	rs, err := r.dao.ListResource(ctx, fields, modelUid, offset, limit)
	if err != nil {
		return nil, err
	}

	res := make([]domain.Resource, 0, len(rs))

	for _, val := range rs {
		res = append(res, domain.Resource{
			ID:       val.ID,
			ModelUID: val.ModelUID,
			Data:     val.Data,
			Name:     val.Name,
		})
	}

	return res, nil
}

func (r *resourceRepository) ListExcludeResource(ctx context.Context, fields []string, modelUid string, offset, limit int64, ids []int64) ([]domain.Resource, error) {
	rs, err := r.dao.ListExcludeResource(ctx, fields, modelUid, offset, limit, ids)
	if err != nil {
		return nil, err
	}

	res := make([]domain.Resource, 0, len(rs))

	for _, val := range rs {
		res = append(res, domain.Resource{
			ID:       val.ID,
			ModelUID: val.ModelUID,
			Data:     val.Data,
			Name:     val.Name,
		})
	}

	return res, nil
}

func (r *resourceRepository) toEntity(req domain.Resource) dao.Resource {
	return dao.Resource{
		ModelUID: req.ModelUID,
		Name:     req.Name,
		Data:     req.Data,
	}
}
