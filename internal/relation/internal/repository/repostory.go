package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"time"
)

type RelationRepository interface {
	CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error)
	CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error)
	ListModelRelation(ctx context.Context, offset, limit int64) ([]domain.ModelRelation, error)
	ListResourceRelation(ctx context.Context, offset, limit int64) ([]domain.ResourceRelation, error)
	Total(ctx context.Context) (int64, error)

	ListRelationByModelIdentifies(ctx context.Context, offset, limit int64, modelUid string) ([]domain.ModelRelation, error)
	TotalByModelIdentifies(ctx context.Context, modelUid string) (int64, error)
}

func NewRelationRepository(dao dao.RelationDAO) RelationRepository {
	return &relationRepository{
		dao: dao,
	}
}

type relationRepository struct {
	dao dao.RelationDAO
}

func (r *relationRepository) CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error) {
	return r.dao.CreateModelRelation(ctx, dao.ModelRelation{
		SourceModelUID:  req.SourceModelUID,
		TargetModelUID:  req.TargetModelUID,
		RelationTypeUID: req.RelationTypeUID,
		Mapping:         req.Mapping,
	})
}

func (r *relationRepository) CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error) {
	return r.dao.CreateResourceRelation(ctx, dao.ResourceRelation{
		SourceModelUID:   req.SourceModelUID,
		TargetModelUID:   req.TargetModelUID,
		RelationTypeUID:  req.RelationTypeUID,
		SourceResourceID: req.SourceResourceID,
		TargetResourceID: req.TargetResourceID,
	})
}

func (r *relationRepository) ListModelRelation(ctx context.Context, offset, limit int64) ([]domain.ModelRelation, error) {
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

func (r *relationRepository) ListResourceRelation(ctx context.Context, offset, limit int64) ([]domain.ResourceRelation, error) {
	resourceRelations, err := r.dao.ListResourceRelation(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	res := make([]domain.ResourceRelation, 0, len(resourceRelations))

	for _, value := range resourceRelations {
		res = append(res, r.toResourceDomain(value))
	}

	return res, nil
}

func (r *relationRepository) ListRelationByModelIdentifies(ctx context.Context, offset, limit int64, modelUid string) ([]domain.ModelRelation, error) {
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

func (r *relationRepository) TotalByModelIdentifies(ctx context.Context, modelUid string) (int64, error) {
	return r.dao.CountByModelUid(ctx, modelUid)
}

func (r *relationRepository) Total(ctx context.Context) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *relationRepository) toResourceDomain(resourceDao *dao.ResourceRelation) domain.ResourceRelation {
	return domain.ResourceRelation{
		ID:               resourceDao.Id,
		SourceModelUID:   resourceDao.SourceModelUID,
		TargetModelUID:   resourceDao.TargetModelUID,
		SourceResourceID: resourceDao.SourceResourceID,
		TargetResourceID: resourceDao.TargetResourceID,
		RelationTypeUID:  resourceDao.RelationTypeUID,
		RelationName:     resourceDao.RelationName,
		Ctime:            time.UnixMilli(resourceDao.Ctime),
		Utime:            time.UnixMilli(resourceDao.Utime),
	}
}

func (r *relationRepository) toDomain(modelDao *dao.ModelRelation) domain.ModelRelation {
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
