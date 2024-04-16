package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"time"
)

type RelationResourceRepository interface {
	CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error)
	ListResourceRelation(ctx context.Context, offset, limit int64) ([]domain.ResourceRelation, error)
	ListResourceIds(ctx context.Context, modelUid string, relationType string) ([]int64, error)
}

func NewRelationResourceRepository(dao dao.RelationResourceDAO) RelationResourceRepository {
	return &relationResourceRepository{
		dao: dao,
	}
}

type relationResourceRepository struct {
	dao dao.RelationResourceDAO
}

func (r *relationResourceRepository) CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error) {
	return r.dao.CreateResourceRelation(ctx, dao.ResourceRelation{
		SourceModelUID:   req.SourceModelUID,
		TargetModelUID:   req.TargetModelUID,
		RelationTypeUID:  req.RelationTypeUID,
		SourceResourceID: req.SourceResourceID,
		TargetResourceID: req.TargetResourceID,
	})
}

func (r *relationResourceRepository) ListResourceRelation(ctx context.Context, offset, limit int64) ([]domain.ResourceRelation, error) {
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

func (r *relationResourceRepository) TotalByModelIdentifies(ctx context.Context, modelUid string) (int64, error) {
	return r.dao.CountByModelUid(ctx, modelUid)
}

func (r *relationResourceRepository) ListResourceIds(ctx context.Context, modelUid string, relationType string) ([]int64, error) {
	return r.dao.ListResourceIds(ctx, modelUid, relationType)
}

func (r *relationResourceRepository) toResourceDomain(resourceDao *dao.ResourceRelation) domain.ResourceRelation {
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
