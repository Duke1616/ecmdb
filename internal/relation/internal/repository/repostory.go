package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
)

type RelationRepository interface {
	CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error)
	CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error)
}

func NewRelationRepository(dao dao.RelationDAO) RelationRepository {
	return &relationRepository{
		dao: dao,
	}
}

type relationRepository struct {
	dao dao.RelationDAO
}

func (r relationRepository) CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error) {
	return r.dao.CreateModelRelation(ctx, dao.ModelRelation{
		SourceModelIdentifies:  req.SourceModelIdentifies,
		TargetModelIdentifies:  req.TargetModelIdentifies,
		RelationTypeIdentifies: req.RelationTypeIdentifies,
		Mapping:                req.Mapping,
	})
}

func (r relationRepository) CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error) {
	return r.dao.CreateResourceRelation(ctx, dao.ResourceRelation{
		SourceModelIdentifies:  req.SourceModelIdentifies,
		TargetModelIdentifies:  req.TargetModelIdentifies,
		RelationTypeIdentifies: req.RelationTypeIdentifies,
		SourceResourceID:       req.SourceResourceID,
		TargetResourceID:       req.TargetResourceID,
	})
}
