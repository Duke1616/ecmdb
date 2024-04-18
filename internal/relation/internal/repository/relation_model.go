package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"time"
)

type RelationModelRepository interface {
	CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error)
	ListModelRelation(ctx context.Context, offset, limit int64) ([]domain.ModelRelation, error)
	Total(ctx context.Context) (int64, error)

	ListRelationByModelUid(ctx context.Context, offset, limit int64, modelUid string) ([]domain.ModelRelation, error)
	TotalByModelUid(ctx context.Context, modelUid string) (int64, error)

	ListSrcModelByUid(ctx context.Context, sourceUid string) ([]domain.ModelRelationDiagram, error)
	ListDstModelByUid(ctx context.Context, sourceUid string) ([]domain.ModelRelationDiagram, error)
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
	return r.dao.CreateModelRelation(ctx, dao.ModelRelation{
		SourceModelUID:  req.SourceModelUID,
		TargetModelUID:  req.TargetModelUID,
		RelationTypeUID: req.RelationTypeUID,
		Mapping:         req.Mapping,
	})
}

func (r *modelRepository) ListModelRelation(ctx context.Context, offset, limit int64) ([]domain.ModelRelation, error) {
	modelRelations, err := r.dao.ListModelRelation(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	res := make([]domain.ModelRelation, 0, len(modelRelations))

	for _, value := range modelRelations {
		res = append(res, r.toDomain(value))
	}

	return res, nil

}

func (r *modelRepository) Total(ctx context.Context) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *modelRepository) ListRelationByModelUid(ctx context.Context, offset, limit int64, modelUid string) ([]domain.ModelRelation, error) {
	relations, err := r.dao.ListRelationByModelUid(ctx, offset, limit, modelUid)
	if err != nil {
		return nil, err
	}

	res := make([]domain.ModelRelation, 0, len(relations))
	for _, value := range relations {
		res = append(res, r.toDomain(value))
	}

	return res, nil
}

func (r *modelRepository) TotalByModelUid(ctx context.Context, modelUid string) (int64, error) {
	return r.dao.CountByModelUid(ctx, modelUid)
}

func (r *modelRepository) ListSrcModelByUid(ctx context.Context, sourceUid string) ([]domain.ModelRelationDiagram, error) {
	mrs, err := r.dao.ListSrcModelByUid(ctx, sourceUid)
	if err != nil {
		return nil, err
	}

	res := make([]domain.ModelRelationDiagram, 0, len(mrs))
	for _, value := range mrs {
		res = append(res, domain.ModelRelationDiagram{
			ID:              value.Id,
			RelationTypeUID: value.RelationTypeUID,
			TargetModelUID:  value.TargetModelUID,
		})
	}

	return res, nil
}

func (r *modelRepository) ListDstModelByUid(ctx context.Context, sourceUid string) ([]domain.ModelRelationDiagram, error) {
	mrs, err := r.dao.ListDstModelByUid(ctx, sourceUid)
	if err != nil {
		return nil, err
	}

	res := make([]domain.ModelRelationDiagram, 0, len(mrs))
	for _, value := range mrs {
		res = append(res, domain.ModelRelationDiagram{
			ID:              value.Id,
			RelationTypeUID: value.RelationTypeUID,
			TargetModelUID:  value.TargetModelUID,
		})
	}

	return res, nil
}

func (r *modelRepository) toDomain(modelDao *dao.ModelRelation) domain.ModelRelation {
	return domain.ModelRelation{
		ID:              modelDao.Id,
		SourceModelUID:  modelDao.SourceModelUID,
		TargetModelUID:  modelDao.TargetModelUID,
		Mapping:         modelDao.Mapping,
		RelationName:    modelDao.RelationName,
		RelationTypeUID: modelDao.RelationTypeUID,
		Ctime:           time.UnixMilli(modelDao.Ctime),
		Utime:           time.UnixMilli(modelDao.Utime),
	}
}
