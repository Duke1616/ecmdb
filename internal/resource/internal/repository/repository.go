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
	TotalByModelUid(ctx context.Context, modelUid string) (int64, error)
	SetCustomField(ctx context.Context, id int64, field string, data interface{}) (int64, error)
	ListResourcesByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error)
	DeleteResource(ctx context.Context, id int64) (int64, error)
	ListExcludeAndFilterResourceByIds(ctx context.Context, fields []string, modelUid string, offset, limit int64,
		ids []int64, filter domain.Condition) ([]domain.Resource, error)
	TotalExcludeAndFilterResourceByIds(ctx context.Context, modelUid string, ids []int64, filter domain.Condition) (int64, error)
	Search(ctx context.Context, text string) ([]domain.SearchResource, error)

	FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error)
	UpdateResource(ctx context.Context, resource domain.Resource) (int64, error)

	CountByModelUids(ctx context.Context, modelUids []string) (map[string]int, error)
}

type resourceRepository struct {
	dao dao.ResourceDAO
}

func (repo *resourceRepository) SetCustomField(ctx context.Context, id int64, field string, data interface{}) (int64, error) {
	return repo.dao.SetCustomField(ctx, id, field, data)
}

func NewResourceRepository(dao dao.ResourceDAO) ResourceRepository {
	return &resourceRepository{
		dao: dao,
	}
}

func (repo *resourceRepository) UpdateResource(ctx context.Context, resource domain.Resource) (int64, error) {
	return repo.dao.UpdateAttribute(ctx, repo.toEntity(resource))
}

func (repo *resourceRepository) CountByModelUids(ctx context.Context, modelUids []string) (map[string]int, error) {
	return repo.dao.CountByModelUids(ctx, modelUids)
}

func (repo *resourceRepository) CreateResource(ctx context.Context, req domain.Resource) (int64, error) {
	return repo.dao.CreateResource(ctx, repo.toEntity(req))
}

func (repo *resourceRepository) FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error) {
	rs, err := repo.dao.FindResourceById(ctx, fields, id)
	return repo.toDomain(rs), err
}

func (repo *resourceRepository) ListResourcesByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error) {
	rrs, err := repo.dao.ListResourcesByIds(ctx, fields, ids)

	return slice.Map(rrs, func(idx int, src dao.Resource) domain.Resource {
		return repo.toDomain(src)
	}), err
}

func (repo *resourceRepository) ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]domain.Resource, error) {
	rrs, err := repo.dao.ListResource(ctx, fields, modelUid, offset, limit)

	return slice.Map(rrs, func(idx int, src dao.Resource) domain.Resource {
		return repo.toDomain(src)
	}), err
}

func (repo *resourceRepository) TotalByModelUid(ctx context.Context, modelUid string) (int64, error) {
	return repo.dao.CountByModelUid(ctx, modelUid)
}

func (repo *resourceRepository) DeleteResource(ctx context.Context, id int64) (int64, error) {
	return repo.dao.DeleteResource(ctx, id)
}

func (repo *resourceRepository) Search(ctx context.Context, text string) ([]domain.SearchResource, error) {
	search, err := repo.dao.Search(ctx, text)

	return slice.Map(search, func(idx int, src dao.SearchResource) domain.SearchResource {
		return domain.SearchResource{
			ModelUid: src.ModelUid,
			Total:    src.Total,
			Data:     src.Data,
		}
	}), err
}

func (repo *resourceRepository) ListExcludeAndFilterResourceByIds(ctx context.Context, fields []string, modelUid string,
	offset, limit int64, ids []int64, filter domain.Condition) ([]domain.Resource, error) {
	rrs, err := repo.dao.ListExcludeAndFilterResourceByIds(ctx, fields, modelUid, offset, limit, ids, filter)

	return slice.Map(rrs, func(idx int, src dao.Resource) domain.Resource {
		return repo.toDomain(src)
	}), err
}

func (repo *resourceRepository) TotalExcludeAndFilterResourceByIds(ctx context.Context, modelUid string, ids []int64,
	filter domain.Condition) (int64, error) {
	return repo.dao.TotalExcludeAndFilterResourceByIds(ctx, modelUid, ids, filter)
}

func (repo *resourceRepository) FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error) {
	return repo.dao.FindSecureData(ctx, id, fieldUid)
}

func (repo *resourceRepository) toEntity(req domain.Resource) dao.Resource {
	return dao.Resource{
		ID:       req.ID,
		ModelUID: req.ModelUID,
		Data:     req.Data,
	}
}

func (repo *resourceRepository) toDomain(src dao.Resource) domain.Resource {
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
