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
	Count(ctx context.Context) (int64, error)
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
		SourceModelIdentifies:  req.SourceModelIdentifies,
		TargetModelIdentifies:  req.TargetModelIdentifies,
		RelationTypeIdentifies: req.RelationTypeIdentifies,
		Mapping:                req.Mapping,
	})
}

func (r *relationRepository) CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error) {
	return r.dao.CreateResourceRelation(ctx, dao.ResourceRelation{
		SourceModelIdentifies:  req.SourceModelIdentifies,
		TargetModelIdentifies:  req.TargetModelIdentifies,
		RelationTypeIdentifies: req.RelationTypeIdentifies,
		SourceResourceID:       req.SourceResourceID,
		TargetResourceID:       req.TargetResourceID,
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

func (r *relationRepository) Count(ctx context.Context) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *relationRepository) toResourceDomain(resourceDao *dao.ResourceRelation) domain.ResourceRelation {
	return domain.ResourceRelation{
		ID:                     resourceDao.Id,
		SourceModelIdentifies:  resourceDao.SourceModelIdentifies,
		TargetModelIdentifies:  resourceDao.TargetModelIdentifies,
		SourceResourceID:       resourceDao.SourceResourceID,
		TargetResourceID:       resourceDao.TargetResourceID,
		RelationTypeIdentifies: resourceDao.RelationTypeIdentifies,
		RelationName:           resourceDao.RelationName,
		Ctime:                  time.UnixMilli(resourceDao.Ctime),
		Utime:                  time.UnixMilli(resourceDao.Utime),
	}
}

func (r *relationRepository) toDomain(modelDao *dao.ModelRelation) domain.ModelRelation {
	return domain.ModelRelation{
		ID:                     modelDao.Id,
		SourceModelIdentifies:  modelDao.SourceModelIdentifies,
		TargetModelIdentifies:  modelDao.TargetModelIdentifies,
		Mapping:                modelDao.Mapping,
		RelationName:           modelDao.RelationName,
		RelationTypeIdentifies: modelDao.RelationTypeIdentifies,
		Ctime:                  time.UnixMilli(modelDao.Ctime),
		Utime:                  time.UnixMilli(modelDao.Utime),
	}
}
