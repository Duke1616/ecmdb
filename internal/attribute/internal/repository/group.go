package repository

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type AttributeGroupRepository interface {
	// CreateAttributeGroup 创建属性组
	CreateAttributeGroup(ctx context.Context, req domain.AttributeGroup) (int64, error)

	// BatchCreateAttributeGroup 批量创建组
	BatchCreateAttributeGroup(ctx context.Context, ags []domain.AttributeGroup) ([]domain.AttributeGroup, error)

	// ListAttributeGroup 根据模型唯一标识，获取组信息
	ListAttributeGroup(ctx context.Context, modelUid string) ([]domain.AttributeGroup, error)

	// ListAttributeGroupByIds 根据 IDS 获取组信息
	ListAttributeGroupByIds(ctx context.Context, ids []int64) ([]domain.AttributeGroup, error)

	// DeleteAttributeGroup 删除属性组
	DeleteAttributeGroup(ctx context.Context, id int64) (int64, error)

	// RenameAttributeGroup 重命名属性组
	RenameAttributeGroup(ctx context.Context, id int64, name string) (int64, error)

	// UpdateSort 更新属性组排序
	UpdateSort(ctx context.Context, id int64, sortKey int64) error

	// BatchUpdateSort 批量更新属性组排序
	BatchUpdateSort(ctx context.Context, items []domain.AttributeGroupSortItem) error
}

type attributeGroupRepository struct {
	dao dao.AttributeGroupDAO
}

func NewAttributeGroupRepository(dao dao.AttributeGroupDAO) AttributeGroupRepository {
	return &attributeGroupRepository{
		dao: dao,
	}
}

func (a *attributeGroupRepository) ListAttributeGroup(ctx context.Context, modelUid string) ([]domain.AttributeGroup, error) {
	ags, err := a.dao.ListAttributeGroup(ctx, modelUid)

	return slice.Map(ags, func(idx int, src dao.AttributeGroup) domain.AttributeGroup {
		return a.toDomain(src)
	}), err
}

func (a *attributeGroupRepository) ListAttributeGroupByIds(ctx context.Context, ids []int64) ([]domain.AttributeGroup, error) {
	ags, err := a.dao.ListAttributeGroupByIds(ctx, ids)

	return slice.Map(ags, func(idx int, src dao.AttributeGroup) domain.AttributeGroup {
		return a.toDomain(src)
	}), err
}

func (a *attributeGroupRepository) CreateAttributeGroup(ctx context.Context, req domain.AttributeGroup) (int64, error) {
	return a.dao.CreateAttributeGroup(ctx, a.toEntity(req))
}

func (a *attributeGroupRepository) BatchCreateAttributeGroup(ctx context.Context, ags []domain.AttributeGroup) ([]domain.AttributeGroup, error) {
	agsResp, err := a.dao.BatchCreateAttributeGroup(ctx, slice.Map(ags, func(idx int, src domain.AttributeGroup) dao.AttributeGroup {
		return a.toEntity(src)
	}))

	return slice.Map(agsResp, func(idx int, src dao.AttributeGroup) domain.AttributeGroup {
		return a.toDomain(src)
	}), err
}

func (a *attributeGroupRepository) toEntity(req domain.AttributeGroup) dao.AttributeGroup {
	return dao.AttributeGroup{
		Name:     req.Name,
		ModelUid: req.ModelUid,
		SortKey:  req.SortKey,
	}
}

func (a *attributeGroupRepository) toDomain(src dao.AttributeGroup) domain.AttributeGroup {
	return domain.AttributeGroup{
		ID:       src.Id,
		Name:     src.Name,
		ModelUid: src.ModelUid,
		SortKey:  src.SortKey,
	}
}

func (a *attributeGroupRepository) DeleteAttributeGroup(ctx context.Context, id int64) (int64, error) {
	return a.dao.DeleteAttributeGroup(ctx, id)
}

func (a *attributeGroupRepository) RenameAttributeGroup(ctx context.Context, id int64, name string) (int64, error) {
	return a.dao.RenameAttributeGroup(ctx, id, name)
}

func (a *attributeGroupRepository) UpdateSort(ctx context.Context, id int64, sortKey int64) error {
	return a.dao.UpdateSort(ctx, id, sortKey)
}

func (a *attributeGroupRepository) BatchUpdateSort(ctx context.Context, items []domain.AttributeGroupSortItem) error {
	daoItems := slice.Map(items, func(idx int, src domain.AttributeGroupSortItem) dao.AttributeGroupSortItem {
		return dao.AttributeGroupSortItem{
			ID:      src.ID,
			SortKey: src.SortKey,
		}
	})
	return a.dao.BatchUpdateSort(ctx, daoItems)
}
