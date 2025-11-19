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
	}
}

func (a *attributeGroupRepository) toDomain(src dao.AttributeGroup) domain.AttributeGroup {
	return domain.AttributeGroup{
		ID:       src.Id,
		Name:     src.Name,
		ModelUid: src.ModelUid,
	}
}
