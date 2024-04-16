package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
)

type RelationTypeRepository interface {
	Create(ctx context.Context, req domain.RelationType) (int64, error)
}

func NewRelationTypeRepository(dao dao.RelationTypeDAO) RelationTypeRepository {
	return &relationRepository{
		dao: dao,
	}
}

type relationRepository struct {
	dao dao.RelationTypeDAO
}

func (r *relationRepository) Create(ctx context.Context, req domain.RelationType) (int64, error) {
	return r.dao.Create(ctx, dao.RelationType{
		UID:            req.UID,
		SourceDescribe: req.SourceDescribe,
		TargetDescribe: req.TargetDescribe,
	})
}
