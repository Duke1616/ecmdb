package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"time"
)

type RelationTypeRepository interface {
	Create(ctx context.Context, req domain.RelationType) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]domain.RelationType, error)
	Total(ctx context.Context) (int64, error)
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
		Name:           req.Name,
		SourceDescribe: req.SourceDescribe,
		TargetDescribe: req.TargetDescribe,
	})
}

func (r *relationRepository) List(ctx context.Context, offset, limit int64) ([]domain.RelationType, error) {
	relations, err := r.dao.List(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	var set []domain.RelationType
	for _, val := range relations {
		set = append(set, domain.RelationType{
			ID:             val.Id,
			UID:            val.UID,
			Name:           val.Name,
			SourceDescribe: val.SourceDescribe,
			TargetDescribe: val.TargetDescribe,
			Ctime:          time.UnixMilli(val.Ctime),
			Utime:          time.UnixMilli(val.Utime),
		})
	}

	return set, nil
}

func (r *relationRepository) Total(ctx context.Context) (int64, error) {
	return r.dao.Count(ctx)
}
