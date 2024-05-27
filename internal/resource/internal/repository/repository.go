package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type ResourceRepository interface {
	CreateResource(ctx context.Context, req domain.Resource) (int64, error)
	FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error)
	ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]domain.Resource, error)
	Total(ctx context.Context, modelUid string) (int64, error)
	ListResourcesByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error)
	DeleteResource(ctx context.Context, id int64) (int64, error)
	ListExcludeAndFilterResourceByIds(ctx context.Context, fields []string, modelUid string, offset, limit int64,
		ids []int64, filter domain.Condition) ([]domain.Resource, error)
	TotalExcludeAndFilterResourceByIds(ctx context.Context, modelUid string, ids []int64, filter domain.Condition) (int64, error)
	PipelineByModelUid(ctx context.Context) (map[string]int, error)
	Search(ctx context.Context, text string) ([]domain.SearchResource, error)
}

type resourceRepository struct {
	dao dao.ResourceDAO
}

func (r *resourceRepository) PipelineByModelUid(ctx context.Context) (map[string]int, error) {
	pipeline, err := r.dao.PipelineByModelUid(ctx)
	if err != nil {
		return nil, err
	}

	p := make(map[string]int, len(pipeline))
	for _, val := range pipeline {
		p[val.ModelUid] = val.Total
	}

	return p, nil
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
	return r.toDomain(rs), err
}

func (r *resourceRepository) ListResourcesByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error) {
	rrs, err := r.dao.ListResourcesByIds(ctx, fields, ids)

	return slice.Map(rrs, func(idx int, src dao.Resource) domain.Resource {
		return r.toDomain(src)
	}), err
}

func (r *resourceRepository) ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]domain.Resource, error) {
	rrs, err := r.dao.ListResource(ctx, fields, modelUid, offset, limit)

	return slice.Map(rrs, func(idx int, src dao.Resource) domain.Resource {
		return r.toDomain(src)
	}), err
}

func (r *resourceRepository) Total(ctx context.Context, modelUid string) (int64, error) {
	return r.dao.Count(ctx, modelUid)
}

func (r *resourceRepository) DeleteResource(ctx context.Context, id int64) (int64, error) {
	return r.dao.DeleteResource(ctx, id)
}

func (r *resourceRepository) Search(ctx context.Context, text string) ([]domain.SearchResource, error) {
	search, err := r.dao.Search(ctx, text)

	return slice.Map(search, func(idx int, src dao.SearchResource) domain.SearchResource {
		return domain.SearchResource{
			ModelUid: src.ModelUid,
			Total:    src.Total,
			Data:     src.Data,
		}
	}), err
}

func (r *resourceRepository) ListExcludeAndFilterResourceByIds(ctx context.Context, fields []string, modelUid string,
	offset, limit int64, ids []int64, filter domain.Condition) ([]domain.Resource, error) {
	rrs, err := r.dao.ListExcludeAndFilterResourceByIds(ctx, fields, modelUid, offset, limit, ids, filter)

	return slice.Map(rrs, func(idx int, src dao.Resource) domain.Resource {
		return r.toDomain(src)
	}), err
}

func (r *resourceRepository) TotalExcludeAndFilterResourceByIds(ctx context.Context, modelUid string, ids []int64,
	filter domain.Condition) (int64, error) {
	return r.dao.TotalExcludeAndFilterResourceByIds(ctx, modelUid, ids, filter)
}

func (r *resourceRepository) toEntity(req domain.Resource) dao.Resource {
	return dao.Resource{
		ModelUID: req.ModelUID,
		Data:     req.Data,
	}
}

func (r *resourceRepository) toDomain(src dao.Resource) domain.Resource {
	name := "undefined"
	if val, ok := src.Data["name"]; ok {
		name = val.(string)
	}

	return domain.Resource{
		ID:       src.ID,
		ModelUID: src.ModelUID,
		Data:     src.Data,
		Name:     name,
	}
}
