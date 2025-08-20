package service

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/rota/internal/domain"
	"github.com/Duke1616/ecmdb/internal/rota/internal/repository"
	"github.com/Duke1616/ecmdb/internal/rota/internal/service/schedule"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	Create(ctx context.Context, req domain.Rota) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]domain.Rota, int64, error)
	Detail(ctx context.Context, id int64) (domain.Rota, error)
	Update(ctx context.Context, rota domain.Rota) (int64, error)
	Delete(ctx context.Context, id int64) (int64, error)

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
	GetCurrentSchedule(ctx context.Context, id int64) (domain.Schedule, error)
}

func NewService(repo repository.RotaRepository, rule schedule.Scheduler) Service {
	return &service{
		repo: repo,
		rule: rule,
	}
}

type service struct {
	rule schedule.Scheduler
	repo repository.RotaRepository
}

func (s *service) GetCurrentSchedule(ctx context.Context, id int64) (domain.Schedule, error) {
	rota, err := s.repo.Detail(ctx, id)
	if err != nil {
		return domain.Schedule{}, err
	}

	if len(rota.Rules) == 0 {
		return domain.Schedule{}, nil
	}

	// TODO 暂时不处理多规则情况，前端控制只能有一条规则
	var sc domain.Schedule
	for _, rule := range rota.Rules {
		sc, err = s.rule.GetCurrentSchedule(rule, rota.AdjustmentRules)
		if err != nil {
			return domain.Schedule{}, err
		}
	}

	return sc, err
}

func (s *service) Delete(ctx context.Context, id int64) (int64, error) {
	return s.repo.Delete(ctx, id)
}

func (s *service) Update(ctx context.Context, rota domain.Rota) (int64, error) {
	return s.repo.Update(ctx, rota)
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

	// TODO 暂时不处理多规则情况，前端控制只能有一条规则
	var rotas []domain.ShiftRostered
	for _, rule := range rota.Rules {
		r, er := s.rule.GenerateSchedule(rule, rota.AdjustmentRules, stime, etime)

		r.Members = toMembers(rule.RotaGroups)
		if er != nil {
			return domain.ShiftRostered{}, er
		}

		rotas = append(rotas, r)
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
