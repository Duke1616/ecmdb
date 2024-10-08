package repository

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type RelationModelRepository interface {
	CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error)
	DeleteModelRelation(ctx context.Context, id int64) (int64, error)
	ListRelationByModelUid(ctx context.Context, offset, limit int64, modelUid string) ([]domain.ModelRelation, error)
	TotalByModelUid(ctx context.Context, modelUid string) (int64, error)

	FindModelDiagramBySrcUids(ctx context.Context, srcUids []string) ([]domain.ModelDiagram, error)
}

func NewRelationModelRepository(dao dao.RelationModelDAO) RelationModelRepository {
	return &modelRepository{
		dao: dao,
	}
}

type modelRepository struct {
	dao dao.RelationModelDAO
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
		SourceModelUid:  req.SourceModelUID,
		TargetModelUid:  req.TargetModelUID,
		RelationName:    fmt.Sprintf("%s_%s_%s", req.SourceModelUID, req.RelationTypeUID, req.TargetModelUID),
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
