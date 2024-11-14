package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/rota/internal/domain"
	"github.com/Duke1616/ecmdb/internal/rota/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type RotaRepository interface {
	Create(ctx context.Context, req domain.Rota) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]domain.Rota, error)
	Total(ctx context.Context) (int64, error)
	AddRole(ctx context.Context, id int64, rr domain.RotaRule) (int64, error)
}

func NewRotaRepository(dao dao.RotaDao) RotaRepository {
	return &rotaRepository{
		dao: dao,
	}
}

type rotaRepository struct {
	dao dao.RotaDao
}

func (repo *rotaRepository) List(ctx context.Context, offset, limit int64) ([]domain.Rota, error) {
	rs, err := repo.dao.List(ctx, offset, limit)
	return slice.Map(rs, func(idx int, src dao.Rota) domain.Rota {
		return repo.toDomain(src)
	}), err
}

func (repo *rotaRepository) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func (repo *rotaRepository) AddRole(ctx context.Context, id int64, rr domain.RotaRule) (int64, error) {
	return repo.dao.FindOrUpdateRole(ctx, id, repo.toRuleEntity(rr))
}

func (repo *rotaRepository) Create(ctx context.Context, req domain.Rota) (int64, error) {
	return repo.dao.Create(ctx, repo.toEntity(req))
}

func (repo *rotaRepository) toEntity(req domain.Rota) dao.Rota {
	return dao.Rota{
		Id:      req.Id,
		Name:    req.Name,
		Desc:    req.Desc,
		Enabled: req.Enabled,
	}
}

func (repo *rotaRepository) toDomain(req dao.Rota) domain.Rota {
	return domain.Rota{
		Id:      req.Id,
		Name:    req.Name,
		Desc:    req.Desc,
		Enabled: req.Enabled,
		Owner:   req.Owner,
		Rules: slice.Map(req.Rules, func(idx int, src dao.RotaRule) domain.RotaRule {
			return repo.toRuleDomain(src)
		}),
		TempRules: slice.Map(req.TempRules, func(idx int, src dao.RotaRule) domain.RotaRule {
			return repo.toRuleDomain(src)
		}),
	}
}

func (repo *rotaRepository) toRuleEntity(req domain.RotaRule) dao.RotaRule {
	return dao.RotaRule{
		RotaGroups: slice.Map(req.RotaGroups, func(idx int, src domain.RotaGroup) dao.RotaGroup {
			return dao.RotaGroup{
				Name:    src.Name,
				Members: src.Members,
			}
		}),
		IsRotate: req.IsRotate,
		Rotate: dao.Rotate{
			TimeUnit:     req.Rotate.TimeUnit.ToUint8(),
			TimeDuration: req.Rotate.TimeDuration,
		},
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
}

func (repo *rotaRepository) toRuleDomain(req dao.RotaRule) domain.RotaRule {
	return domain.RotaRule{
		RotaGroups: slice.Map(req.RotaGroups, func(idx int, src dao.RotaGroup) domain.RotaGroup {
			return domain.RotaGroup{
				Name:    src.Name,
				Members: src.Members,
			}
		}),
		IsRotate: req.IsRotate,
		Rotate: domain.Rotate{
			TimeUnit:     domain.TimeUnit(req.Rotate.TimeUnit),
			TimeDuration: req.Rotate.TimeDuration,
		},
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
}
