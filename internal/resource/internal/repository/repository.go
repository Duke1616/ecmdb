package repository

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

//go:generate mockgen -source=repository.go -destination=../../mocks/repository.mock.go --package=resourcemocks ResourceRepository
type ResourceRepository interface {
	// CreateResource 创建资产
	CreateResource(ctx context.Context, req domain.Resource) (int64, error)

	// FindResourceById 根据 ID 获取资产信息
	FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error)

	// ListResource 获取指定模型的资产列表
	ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]domain.Resource, error)

	// TotalByModelUid 获取指定模型的资产总数
	TotalByModelUid(ctx context.Context, modelUid string) (int64, error)

	// SetCustomField 设置指定资产的自定义字段值
	SetCustomField(ctx context.Context, id int64, field string, data interface{}) (int64, error)

	// ListResourcesByIds 根据 ID 列表批量获取资产
	ListResourcesByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error)

	// DeleteResource 删除指定资产
	DeleteResource(ctx context.Context, id int64) (int64, error)

	// ListExcludeAndFilterResourceByIds 排除指定 ID 并根据条件过滤资产列表
	ListExcludeAndFilterResourceByIds(ctx context.Context, fields []string, modelUid string, offset, limit int64,
		ids []int64, filter domain.Condition) ([]domain.Resource, error)

	// TotalExcludeAndFilterResourceByIds 排除指定 ID 并根据条件统计资产总数
	TotalExcludeAndFilterResourceByIds(ctx context.Context, modelUid string, ids []int64, filter domain.Condition) (int64, error)

	// Search 全局搜索资产
	Search(ctx context.Context, text string) ([]domain.SearchResource, error)

	// FindSecureData 查找指定资产的加密字段数据
	FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error)

	// UpdateResource 更新资产数据
	UpdateResource(ctx context.Context, resource domain.Resource) (int64, error)

	// CountByModelUids 统计多个模型的资产数量
	CountByModelUids(ctx context.Context, modelUids []string) (map[string]int, error)

	// BatchUpdateResources 批量更新资产
	BatchUpdateResources(ctx context.Context, resources []domain.Resource) (int64, error)

	// BatchCreateOrUpdate 批量创建或更新资产
	// 基于 model_uid + name 进行 upsert,name 已存在则更新,不存在则创建
	// NOTE: 使用 MongoDB BulkWrite 提升性能,适用于 Excel 导入等批量操作场景
	BatchCreateOrUpdate(ctx context.Context, resources []domain.Resource) error

	// ListBeforeUtime 获取指定时间前的资产列表
	ListBeforeUtime(ctx context.Context, utime int64, fields []string, modelUid string,
		offset, limit int64) ([]domain.Resource, error)
}

type resourceRepository struct {
	dao dao.ResourceDAO
}

func (repo *resourceRepository) ListBeforeUtime(ctx context.Context, utime int64, fields []string, modelUid string,
	offset, limit int64) ([]domain.Resource, error) {
	rrs, err := repo.dao.ListBeforeUtime(ctx, utime, fields, modelUid, offset, limit)

	return slice.Map(rrs, func(idx int, src dao.Resource) domain.Resource {
		return repo.toDomain(src)
	}), err
}

func (repo *resourceRepository) BatchUpdateResources(ctx context.Context, resources []domain.Resource) (int64, error) {
	return repo.dao.BatchUpdateResources(ctx, slice.Map(resources, func(idx int, src domain.Resource) dao.Resource {
		return repo.toEntity(src)
	}))
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

// BatchCreateOrUpdate 批量创建或更新资产
func (repo *resourceRepository) BatchCreateOrUpdate(ctx context.Context, resources []domain.Resource) error {
	return repo.dao.BatchCreateOrUpdate(ctx, slice.Map(resources, func(idx int, src domain.Resource) dao.Resource {
		return repo.toEntity(src)
	}))
}
