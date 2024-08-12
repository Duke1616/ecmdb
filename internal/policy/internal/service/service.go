package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/policy/internal/domain"
	"github.com/casbin/casbin/v2"
	"github.com/ecodeclub/ekit/slice"
	"strconv"
)

type Service interface {
	AddPolicies(ctx context.Context, req domain.Policies) (bool, error)
	AddGroupingPolicy(ctx context.Context, req domain.AddGroupingPolicy) (bool, error)
	CreateOrUpdateFilteredPolicies(ctx context.Context, req domain.Policies) (bool, error)
	Authorize(ctx context.Context, userId, path, method string) (bool, error)
	GetImplicitPermissionsForUser(ctx context.Context, userId string) ([]domain.Policy, error)
	GetPermissionsForRole(ctx context.Context, roleCode string) ([]domain.Policy, error)
	UpdateFilteredGrouping(ctx context.Context, userId int64, roleCodes []string) (bool, error)
}

type service struct {
	enforcer *casbin.SyncedEnforcer
}

func (s *service) GetPermissionsForRole(ctx context.Context, roleCode string) ([]domain.Policy, error) {
	pers, err := s.enforcer.GetPermissionsForUser(roleCode)

	return slice.Map(pers, func(idx int, src []string) domain.Policy {
		return domain.Policy{
			Path:   src[1],
			Method: src[2],
			Effect: domain.Effect(src[3]),
		}
	}), err

}

func (s *service) GetImplicitPermissionsForUser(ctx context.Context, userId string) ([]domain.Policy, error) {
	pers, err := s.enforcer.GetImplicitPermissionsForUser(userId)
	uniquePermissions := make(map[domain.Policy]bool)
	return slice.FilterMap(pers, func(idx int, src []string) (domain.Policy, bool) {
		policy := domain.Policy{
			Path:   src[1],
			Method: src[2],
			Effect: domain.Effect(src[3]),
		}

		if !uniquePermissions[policy] {
			uniquePermissions[policy] = true
			return policy, true
		}

		return policy, false
	}), err
}

func (s *service) Authorize(ctx context.Context, userId, path, method string) (bool, error) {
	//s.enforcer.GetPermissionsForUser()
	return s.enforcer.Enforce(userId, path, method)
}

func (s *service) AddPolicies(ctx context.Context, req domain.Policies) (bool, error) {
	var policies [][]string
	for _, policy := range req.Policies {
		policies = append(policies, []string{req.RoleCode, policy.Path, policy.Method, policy.Effect.ToString()})
	}

	ok, err := s.enforcer.AddPolicies(policies)
	if err != nil {
		return ok, err
	}
	return ok, nil
}

func (s *service) CreateOrUpdateFilteredPolicies(ctx context.Context, req domain.Policies) (bool, error) {
	var policies [][]string
	for _, policy := range req.Policies {
		policies = append(policies, []string{req.RoleCode, policy.Path, policy.Method, policy.Effect.ToString()})
	}

	ok, err := s.enforcer.UpdateFilteredPolicies(policies, 0, req.RoleCode)
	if err != nil {
		return ok, err
	}

	return ok, nil
}

func (s *service) AddGroupingPolicy(ctx context.Context, req domain.AddGroupingPolicy) (bool, error) {
	ok, err := s.enforcer.AddGroupingPolicy(req.UserId, req.RoleCode)
	if err != nil {
		return ok, err
	}
	return ok, nil
}

func (s *service) UpdateFilteredGrouping(ctx context.Context, userId int64, roleCodes []string) (bool, error) {
	var rcs [][]string
	for _, policy := range roleCodes {
		rcs = append(rcs, []string{strconv.FormatInt(userId, 10), policy})
	}

	ok, err := s.enforcer.RemoveFilteredGroupingPolicy(0, strconv.FormatInt(userId, 10))
	if err != nil {
		return ok, err
	}

	if rcs == nil {
		return true, nil
	}

	return s.enforcer.AddGroupingPolicies(rcs)
}

func NewService(enforcer *casbin.SyncedEnforcer) Service {
	return &service{
		enforcer: enforcer,
	}
}
