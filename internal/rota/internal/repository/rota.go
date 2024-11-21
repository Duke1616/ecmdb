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
	Update(ctx context.Context, req domain.Rota) (int64, error)
	Detail(ctx context.Context, id int64) (domain.Rota, error)
	Delete(ctx context.Context, id int64) (int64, error)

	AddSchedulingRule(ctx context.Context, id int64, rr domain.RotaRule) (int64, error)
	UpdateSchedulingRule(ctx context.Context, id int64, rotaRules []domain.RotaRule) (int64, error)

	AddAdjustmentRule(ctx context.Context, id int64, rr domain.RotaAdjustmentRule) (int64, error)
	UpdateAdjustmentRule(ctx context.Context, id int64, rotaRules []domain.RotaAdjustmentRule) (int64, error)
}

func NewRotaRepository(dao dao.RotaDao) RotaRepository {
	return &rotaRepository{
		dao: dao,
	}
}

type rotaRepository struct {
	dao dao.RotaDao
}

func (repo *rotaRepository) Delete(ctx context.Context, id int64) (int64, error) {
	return repo.dao.Delete(ctx, id)
}

func (repo *rotaRepository) Update(ctx context.Context, req domain.Rota) (int64, error) {
	return repo.dao.Update(ctx, repo.toEntity(req))
}

func (repo *rotaRepository) UpdateAdjustmentRule(ctx context.Context, id int64, rotaRules []domain.RotaAdjustmentRule) (int64, error) {
	return repo.dao.UpdateAdjustmentRule(ctx, id,
		slice.Map(rotaRules, func(idx int, src domain.RotaAdjustmentRule) dao.RotaAdjustmentRule {
			return repo.toAdjustmentRuleEntity(src)
		}))
}

func (repo *rotaRepository) AddAdjustmentRule(ctx context.Context, id int64, rr domain.RotaAdjustmentRule) (int64, error) {
	return repo.dao.FindOrAddAdjustmentRule(ctx, id, repo.toAdjustmentRuleEntity(rr))
}

func (repo *rotaRepository) UpdateSchedulingRule(ctx context.Context, id int64, rotaRules []domain.RotaRule) (int64, error) {
	return repo.dao.UpdateSchedulingRule(ctx, id, slice.Map(rotaRules, func(idx int, src domain.RotaRule) dao.RotaRule {
		return repo.toRuleEntity(src)
	}))
}

func (repo *rotaRepository) Detail(ctx context.Context, id int64) (domain.Rota, error) {
	rota, err := repo.dao.Detail(ctx, id)
	return repo.toDomain(rota), err
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

func (repo *rotaRepository) AddSchedulingRule(ctx context.Context, id int64, rr domain.RotaRule) (int64, error) {
	return repo.dao.FindOrAddSchedulingRule(ctx, id, repo.toRuleEntity(rr))
}

func (repo *rotaRepository) Create(ctx context.Context, req domain.Rota) (int64, error) {
	return repo.dao.Create(ctx, repo.toEntity(req))
}

func (repo *rotaRepository) toEntity(req domain.Rota) dao.Rota {
	return dao.Rota{
		Id:              req.Id,
		Name:            req.Name,
		Desc:            req.Desc,
		Owner:           req.Owner,
		Enabled:         req.Enabled,
		Rules:           []dao.RotaRule{},
		AdjustmentRules: []dao.RotaAdjustmentRule{},
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
		AdjustmentRules: slice.Map(req.AdjustmentRules, func(idx int, src dao.RotaAdjustmentRule) domain.RotaAdjustmentRule {
			return repo.toAdjustmentRuleDomain(src)
		}),
	}
}

func (repo *rotaRepository) toRuleEntity(req domain.RotaRule) dao.RotaRule {
	return dao.RotaRule{
		RotaGroups: slice.Map(req.RotaGroups, func(idx int, src domain.RotaGroup) dao.RotaGroup {
			return dao.RotaGroup{
				Id:      src.Id,
				Name:    src.Name,
				Members: src.Members,
			}
		}),
		Rotate: dao.Rotate{
			TimeUnit:     req.Rotate.TimeUnit.ToUint8(),
			TimeDuration: req.Rotate.TimeDuration,
		},
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
}

func (repo *rotaRepository) toAdjustmentRuleEntity(req domain.RotaAdjustmentRule) dao.RotaAdjustmentRule {
	return dao.RotaAdjustmentRule{
		RotaGroup: dao.RotaGroup{
			Id:      req.RotaGroup.Id,
			Name:    req.RotaGroup.Name,
			Members: req.RotaGroup.Members,
		},
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
}

func (repo *rotaRepository) toRuleDomain(req dao.RotaRule) domain.RotaRule {
	return domain.RotaRule{
		RotaGroups: slice.Map(req.RotaGroups, func(idx int, src dao.RotaGroup) domain.RotaGroup {
			return domain.RotaGroup{
				Id:      src.Id,
				Name:    src.Name,
				Members: src.Members,
			}
		}),
		Rotate: domain.Rotate{
			TimeUnit:     domain.TimeUnit(req.Rotate.TimeUnit),
			TimeDuration: req.Rotate.TimeDuration,
		},
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
}

func (repo *rotaRepository) toAdjustmentRuleDomain(req dao.RotaAdjustmentRule) domain.RotaAdjustmentRule {
	return domain.RotaAdjustmentRule{
		RotaGroup: domain.RotaGroup{
			Id:      req.RotaGroup.Id,
			Name:    req.RotaGroup.Name,
			Members: req.RotaGroup.Members,
		},
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
}
