package repository

import (
	"context"
	"time"

	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type RelationTypeRepository interface {
	// Create 创建关联类型
	Create(ctx context.Context, req domain.RelationType) (int64, error)

	// BatchCreate 批量创建
	BatchCreate(ctx context.Context, rts []domain.RelationType) error

	// GetByUids 根据 UID 获取关联类型
	GetByUids(ctx context.Context, uids []string) ([]domain.RelationType, error)

	// List 关联列表
	List(ctx context.Context, offset, limit int64) ([]domain.RelationType, error)

	// Total 数量
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

func (r *relationRepository) GetByUids(ctx context.Context, uids []string) ([]domain.RelationType, error) {
	rts, err := r.dao.GetByUids(ctx, uids)
	return slice.Map(rts, func(idx int, src dao.RelationType) domain.RelationType {
		return r.toDomain(src)
	}), err
}

func (r *relationRepository) BatchCreate(ctx context.Context, rts []domain.RelationType) error {
	return r.dao.BatchCreate(ctx, slice.Map(rts, func(idx int, src domain.RelationType) dao.RelationType {
		return r.toEntity(src)
	}))
}

func (r *relationRepository) Create(ctx context.Context, req domain.RelationType) (int64, error) {
	return r.dao.Create(ctx, r.toEntity(req))
}

func (r *relationRepository) List(ctx context.Context, offset, limit int64) ([]domain.RelationType, error) {
	rts, err := r.dao.List(ctx, offset, limit)

	return slice.Map(rts, func(idx int, src dao.RelationType) domain.RelationType {
		return r.toDomain(src)
	}), err
}

func (r *relationRepository) Total(ctx context.Context) (int64, error) {
	return r.dao.Count(ctx)
}

func (r *relationRepository) toEntity(req domain.RelationType) dao.RelationType {
	return dao.RelationType{
		UID:            req.UID,
		Name:           req.Name,
		SourceDescribe: req.SourceDescribe,
		TargetDescribe: req.TargetDescribe,
	}
}

func (r *relationRepository) toDomain(src dao.RelationType) domain.RelationType {
	return domain.RelationType{
		ID:             src.Id,
		UID:            src.UID,
		Name:           src.Name,
		SourceDescribe: src.SourceDescribe,
		TargetDescribe: src.TargetDescribe,
		Ctime:          time.UnixMilli(src.Ctime),
		Utime:          time.UnixMilli(src.Utime),
	}
}
