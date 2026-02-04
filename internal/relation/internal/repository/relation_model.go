package repository

import (
	"context"
	"time"

	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type RelationModelRepository interface {
	// CreateModelRelation 创建模型关联关系
	CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error)

	// DeleteModelRelation 删除模型关联关系
	DeleteModelRelation(ctx context.Context, id int64) (int64, error)

	// BatchCreate 批量创建模型关联关系
	BatchCreate(ctx context.Context, relations []domain.ModelRelation) error

	// GetByRelationNames 根据唯一标识获取数据
	GetByRelationNames(ctx context.Context, names []string) ([]domain.ModelRelation, error)

	// ListRelationByModelUid 根据模型 UID 获取。支持分页
	ListRelationByModelUid(ctx context.Context, offset, limit int64, modelUid string) ([]domain.ModelRelation, error)

	// TotalByModelUid 根据模型 UID 获取数量
	TotalByModelUid(ctx context.Context, modelUid string) (int64, error)

	// FindModelDiagramBySrcUids 查询模型关联关系，绘制拓扑图
	FindModelDiagramBySrcUids(ctx context.Context, srcUids []string) ([]domain.ModelDiagram, error)

	// CountByRelationTypeUID 根据关联类型 UID 获取数量
	CountByRelationTypeUID(ctx context.Context, uid string) (int64, error)

	// GetByID 根据 ID 获取数据
	GetByID(ctx context.Context, id int64) (domain.ModelRelation, error)

	// UpdateModelRelation 更新模型关联关系
	UpdateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error)
}

func NewRelationModelRepository(dao dao.RelationModelDAO) RelationModelRepository {
	return &modelRepository{
		dao: dao,
	}
}

type modelRepository struct {
	dao dao.RelationModelDAO
}

func (r *modelRepository) BatchCreate(ctx context.Context, relations []domain.ModelRelation) error {
	return r.dao.BatchCreate(ctx, slice.Map(relations, func(idx int, src domain.ModelRelation) dao.ModelRelation {
		return r.toEntity(src)
	}))
}

func (r *modelRepository) GetByRelationNames(ctx context.Context, names []string) ([]domain.ModelRelation, error) {
	rms, err := r.dao.GetByRelationNames(ctx, names)

	return slice.Map(rms, func(idx int, src dao.ModelRelation) domain.ModelRelation {
		return r.toDomain(src)
	}), err
}

func (r *modelRepository) CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error) {
	return r.dao.CreateModelRelation(ctx, r.toEntity(req))
}

func (r *modelRepository) ListRelationByModelUid(ctx context.Context, offset, limit int64, modelUid string) ([]domain.ModelRelation, error) {
	rms, err := r.dao.ListRelationByModelUid(ctx, offset, limit, modelUid)

	return slice.Map(rms, func(idx int, src dao.ModelRelation) domain.ModelRelation {
		return r.toDomain(src)
	}), err
}

func (r *modelRepository) TotalByModelUid(ctx context.Context, modelUid string) (int64, error) {
	return r.dao.CountByModelUid(ctx, modelUid)
}

func (r *modelRepository) FindModelDiagramBySrcUids(ctx context.Context, srcUids []string) ([]domain.ModelDiagram, error) {
	diagrams, err := r.dao.FindModelDiagramBySrcUids(ctx, srcUids)

	return slice.Map(diagrams, func(idx int, src dao.ModelRelation) domain.ModelDiagram {
		return r.toDiagram(src)
	}), err
}

func (r *modelRepository) DeleteModelRelation(ctx context.Context, id int64) (int64, error) {
	return r.dao.DeleteModelRelation(ctx, id)
}

func (r *modelRepository) toEntity(req domain.ModelRelation) dao.ModelRelation {
	return dao.ModelRelation{
		Id:              req.ID,
		SourceModelUid:  req.SourceModelUID,
		TargetModelUid:  req.TargetModelUID,
		RelationName:    req.RelationName,
		RelationTypeUid: req.RelationTypeUID,
		Mapping:         req.Mapping,
	}
}

func (r *modelRepository) toDiagram(src dao.ModelRelation) domain.ModelDiagram {
	return domain.ModelDiagram{
		ID:              src.Id,
		RelationTypeUid: src.RelationTypeUid,
		TargetModelUid:  src.TargetModelUid,
		SourceModelUid:  src.SourceModelUid,
	}
}

func (r *modelRepository) toDomain(modelDao dao.ModelRelation) domain.ModelRelation {
	return domain.ModelRelation{
		ID:              modelDao.Id,
		SourceModelUID:  modelDao.SourceModelUid,
		TargetModelUID:  modelDao.TargetModelUid,
		Mapping:         modelDao.Mapping,
		RelationName:    modelDao.RelationName,
		RelationTypeUID: modelDao.RelationTypeUid,
		Ctime:           time.UnixMilli(modelDao.Ctime),
		Utime:           time.UnixMilli(modelDao.Utime),
	}
}

func (r *modelRepository) CountByRelationTypeUID(ctx context.Context, uid string) (int64, error) {
	return r.dao.CountByRelationTypeUid(ctx, uid)
}

func (r *modelRepository) GetByID(ctx context.Context, id int64) (domain.ModelRelation, error) {
	val, err := r.dao.GetByID(ctx, id)
	return r.toDomain(val), err
}

func (r *modelRepository) UpdateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error) {
	return r.dao.UpdateModelRelation(ctx, r.toEntity(req))
}
