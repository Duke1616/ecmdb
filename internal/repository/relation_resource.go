package repository

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

// RelationResourceRepository 资源实例关联关系仓储接口
type RelationResourceRepository interface {
	// CreateResourceRelation 创建资源关联关系
	CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error)

	// ListSrcResources 查询资源列表
	ListSrcResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, error)
	// ListDstResources 根据目标端查询资源关联列表
	ListDstResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, error)
	// TotalSrc 根据源端统计关联数量
	TotalSrc(ctx context.Context, modelUid string, id int64) (int64, error)
	// TotalDst 根据目标端统计关联数量
	TotalDst(ctx context.Context, modelUid string, id int64) (int64, error)

	// ListSrcAggregated 聚合查询关联列表
	ListSrcAggregated(ctx context.Context, modelUid string, id int64) ([]domain.ResourceAggregatedAssets, error)
	// ListDstAggregated 聚合查询目标关联列表
	ListDstAggregated(ctx context.Context, modelUid string, id int64) ([]domain.ResourceAggregatedAssets, error)

	// ListSrcRelated 查询当前已经关联的数据，新增资源关联使用
	ListSrcRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error)
	// ListDstRelated 查询当前已经关联的目标数据，新增资源关联使用
	ListDstRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error)

	// CountByRelationTypeUID 根据关联类型 UID 获取数量
	CountByRelationTypeUID(ctx context.Context, uid string) (int64, error)

	// CountByRelationName 根据关联名称获取数量
	CountByRelationName(ctx context.Context, name string) (int64, error)

	// DeleteResourceRelation 删除资源关联关系
	DeleteResourceRelation(ctx context.Context, id int64) (int64, error)
	// DeleteSrcRelation 删除源端关系
	DeleteSrcRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error)
	// DeleteDstRelation 删除目标端关系
	DeleteDstRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error)

	// ListRecursiveSrc 递归查询下游关联资产列表（正向递归）
	ListRecursiveSrc(ctx context.Context, modelUid string, id int64, maxDepth int) ([]domain.ResourceRelation, error)
	// ListRecursiveDst 递归查询上游关联资产列表（反向递归）
	ListRecursiveDst(ctx context.Context, modelUid string, id int64, maxDepth int) ([]domain.ResourceRelation, error)
}

func NewRelationResourceRepository(dao dao.RelationResourceDAO) RelationResourceRepository {
	return &resourceRelationRepository{
		dao: dao,
	}
}

type resourceRelationRepository struct {
	dao dao.RelationResourceDAO
}

func (r *resourceRelationRepository) CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error) {
	return r.dao.CreateResourceRelation(ctx, r.toEntity(req))
}

func (r *resourceRelationRepository) ListSrcAggregated(ctx context.Context, modelUid string, id int64) ([]domain.ResourceAggregatedAssets, error) {
	rrs, err := r.dao.ListSrcAggregated(ctx, modelUid, id)
	return slice.Map(rrs, func(idx int, src dao.ResourceAggregatedAsset) domain.ResourceAggregatedAssets {
		return r.toAggregatedAssetsDomain(src)
	}), err
}

func (r *resourceRelationRepository) ListDstAggregated(ctx context.Context, modelUid string, id int64) ([]domain.ResourceAggregatedAssets, error) {
	rrs, err := r.dao.ListDstAggregated(ctx, modelUid, id)
	return slice.Map(rrs, func(idx int, src dao.ResourceAggregatedAsset) domain.ResourceAggregatedAssets {
		return r.toAggregatedAssetsDomain(src)
	}), err
}

func (r *resourceRelationRepository) ListSrcResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, error) {
	rrs, err := r.dao.ListSrcResources(ctx, modelUid, id)
	return slice.Map(rrs, func(idx int, src dao.ResourceRelation) domain.ResourceRelation {
		return r.toResourceDomain(src)
	}), err
}

func (r *resourceRelationRepository) ListDstResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, error) {
	rrs, err := r.dao.ListDstResources(ctx, modelUid, id)
	return slice.Map(rrs, func(idx int, src dao.ResourceRelation) domain.ResourceRelation {
		return r.toResourceDomain(src)
	}), err
}

func (r *resourceRelationRepository) TotalSrc(ctx context.Context, modelUid string, id int64) (int64, error) {
	return r.dao.CountSrc(ctx, modelUid, id)
}

func (r *resourceRelationRepository) TotalDst(ctx context.Context, modelUid string, id int64) (int64, error) {
	return r.dao.CountDst(ctx, modelUid, id)
}

func (r *resourceRelationRepository) ListSrcRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error) {
	return r.dao.ListSrcRelated(ctx, modelUid, relationName, id)
}

func (r *resourceRelationRepository) ListDstRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error) {
	return r.dao.ListDstRelated(ctx, modelUid, relationName, id)
}

func (r *resourceRelationRepository) DeleteResourceRelation(ctx context.Context, id int64) (int64, error) {
	return r.dao.DeleteResourceRelation(ctx, id)
}

func (r *resourceRelationRepository) DeleteSrcRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error) {
	return r.dao.DeleteSrcRelation(ctx, resourceId, modelUid, relationName)
}

func (r *resourceRelationRepository) DeleteDstRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error) {
	return r.dao.DeleteDstRelation(ctx, resourceId, modelUid, relationName)
}

func (r *resourceRelationRepository) toEntity(req domain.ResourceRelation) dao.ResourceRelation {
	return dao.ResourceRelation{
		RelationName:     req.RelationName,
		SourceResourceID: req.SourceResourceID,
		TargetResourceID: req.TargetResourceID,
		SourceModelUID:   req.SourceModelUID,
		TargetModelUID:   req.TargetModelUID,
		RelationTypeUID:  req.RelationTypeUID,
	}
}

func (r *resourceRelationRepository) toResourceDomain(resourceDao dao.ResourceRelation) domain.ResourceRelation {
	return domain.ResourceRelation{
		ID:               resourceDao.Id,
		SourceModelUID:   resourceDao.SourceModelUID,
		TargetModelUID:   resourceDao.TargetModelUID,
		SourceResourceID: resourceDao.SourceResourceID,
		TargetResourceID: resourceDao.TargetResourceID,
		RelationTypeUID:  resourceDao.RelationTypeUID,
		RelationName:     resourceDao.RelationName,
	}
}

func (r *resourceRelationRepository) toAggregatedAssetsDomain(src dao.ResourceAggregatedAsset) domain.ResourceAggregatedAssets {
	return domain.ResourceAggregatedAssets{
		RelationName: src.RelationName,
		ModelUid:     src.ModelUid,
		Total:        src.Total,
		ResourceIds:  src.ResourceIds,
	}
}

func (r *resourceRelationRepository) CountByRelationTypeUID(ctx context.Context, uid string) (int64, error) {
	return r.dao.CountByRelationTypeUid(ctx, uid)
}

func (r *resourceRelationRepository) CountByRelationName(ctx context.Context, name string) (int64, error) {
	return r.dao.CountByRelationName(ctx, name)
}

func (r *resourceRelationRepository) ListRecursiveSrc(ctx context.Context, modelUid string, id int64, maxDepth int) ([]domain.ResourceRelation, error) {
	rrs, err := r.dao.ListRecursiveSrc(ctx, modelUid, id, maxDepth)
	return slice.Map(rrs, func(idx int, src dao.ResourceRelation) domain.ResourceRelation {
		return r.toResourceDomain(src)
	}), err
}

func (r *resourceRelationRepository) ListRecursiveDst(ctx context.Context, modelUid string, id int64, maxDepth int) ([]domain.ResourceRelation, error) {
	rrs, err := r.dao.ListRecursiveDst(ctx, modelUid, id, maxDepth)
	return slice.Map(rrs, func(idx int, src dao.ResourceRelation) domain.ResourceRelation {
		return r.toResourceDomain(src)
	}), err
}
