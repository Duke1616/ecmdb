package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/rota/internal/domain"
	"github.com/Duke1616/ecmdb/internal/rota/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	Create(ctx context.Context, req domain.Rota) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]domain.Rota, int64, error)
	Detail(ctx context.Context, id int64) (domain.Rota, error)

	// AddSchedulingRule 常规规则
	AddSchedulingRule(ctx context.Context, id int64, rr domain.RotaRule) (int64, error)
	UpdateSchedulingRule(ctx context.Context, id int64, rotaRules []domain.RotaRule) (int64, error)

	// AddAdjustmentRule 临时规则
	AddAdjustmentRule(ctx context.Context, id int64, rr domain.RotaAdjustmentRule) (int64, error)
	UpdateAdjustmentRule(ctx context.Context, id int64, groupId int64,
		rotaRules domain.RotaAdjustmentRule) (int64, error)
	DeleteAdjustmentRule(ctx context.Context, id int64, groupId int64) (int64, error)

	// GenerateShiftRostered 生成排班表
	GenerateShiftRostered(ctx context.Context, id, stime, etime int64) (domain.ShiftRostered, error)
}

func NewService(repo repository.RotaRepository) Service {
	return &service{
		repo: repo,
	}
}

type service struct {
	repo repository.RotaRepository
}

func (s *service) DeleteAdjustmentRule(ctx context.Context, id int64, groupId int64) (int64, error) {
	rota, err := s.repo.Detail(ctx, id)
	if err != nil {
		return 0, err
	}

	rules := rota.AdjustmentRules[:]
	var newRules []domain.RotaAdjustmentRule
	// 遍历切片并删除目标元素
	for i := 0; i < len(rules); {
		if rules[i].RotaGroup.Id == groupId {
			// 删除当前元素
			newRules = append(rules[:i], rules[i+1:]...)
			break
		} else {
			i++
		}
	}

	return s.repo.UpdateAdjustmentRule(ctx, id, newRules)
}

func (s *service) UpdateAdjustmentRule(ctx context.Context, id int64, groupId int64,
	rotaRule domain.RotaAdjustmentRule) (int64, error) {
	rota, err := s.repo.Detail(ctx, id)
	if err != nil {
		return 0, err
	}

	rules := rota.AdjustmentRules[:] // 将数组转换为切片

	// 遍历切片并删除目标元素
	for i := 0; i < len(rules); {
		if rules[i].RotaGroup.Id == groupId {
			// 删除当前元素
			rules[i] = rotaRule
			break
		} else {
			i++
		}
	}

	return s.repo.UpdateAdjustmentRule(ctx, id, rules)
}

func (s *service) AddAdjustmentRule(ctx context.Context, id int64, rr domain.RotaAdjustmentRule) (int64, error) {
	return s.repo.AddAdjustmentRule(ctx, id, rr)
}

func (s *service) GenerateShiftRostered(ctx context.Context, id, stime, etime int64) (domain.ShiftRostered, error) {
	rota, err := s.repo.Detail(ctx, id)
	if err != nil {
		return domain.ShiftRostered{}, err
	}

	if len(rota.Rules) == 0 {
		return domain.ShiftRostered{}, nil
	}

	var rotas []domain.ShiftRostered
	for _, rule := range rota.Rules {
		r, er := RruleSchedule(rule, stime, etime)

		r.Members = toMembers(rule.RotaGroups)
		if er != nil {
			return domain.ShiftRostered{}, er
		}

		// 处理临时值班插入
		schedule, er := RruleAdjustmentSchedule(r, rota.AdjustmentRules)
		if er != nil {
			return domain.ShiftRostered{}, er
		}

		rotas = append(rotas, schedule)
	}

	return rotas[0], err
}

// 获取用户信息
func toMembers(rotaGroup []domain.RotaGroup) []int64 {
	members := make([]int64, 0)
	for _, group := range rotaGroup {
		seen := make(map[int64]struct{})
		for _, member := range group.Members {
			if _, exists := seen[member]; !exists {
				seen[member] = struct{}{}
				members = append(members, member)
			}
		}
	}

	return members
}

func (s *service) UpdateSchedulingRule(ctx context.Context, id int64, rotaRules []domain.RotaRule) (int64, error) {
	return s.repo.UpdateSchedulingRule(ctx, id, rotaRules)
}

func (s *service) Detail(ctx context.Context, id int64) (domain.Rota, error) {
	return s.repo.Detail(ctx, id)
}

func (s *service) List(ctx context.Context, offset, limit int64) ([]domain.Rota, int64, error) {
	var (
		eg    errgroup.Group
		rs    []domain.Rota
		total int64
	)
	eg.Go(func() error {
		var err error
		rs, err = s.repo.List(ctx, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx)
		return err
	})
	if err := eg.Wait(); err != nil {
		return rs, total, err
	}
	return rs, total, nil
}

func (s *service) Create(ctx context.Context, req domain.Rota) (int64, error) {
	return s.repo.Create(ctx, req)
}

func (s *service) AddSchedulingRule(ctx context.Context, id int64, rr domain.RotaRule) (int64, error) {
	return s.repo.AddSchedulingRule(ctx, id, rr)
}
