package service

import (
	"context"
	"strconv"

	"github.com/Duke1616/ecmdb/internal/policy/internal/domain"
	"github.com/casbin/casbin/v2"
	"github.com/ecodeclub/ekit/slice"
)

type Service interface {
	// AddPolicies 批量添加策略，存在的会跳过
	AddPolicies(ctx context.Context, req domain.Policies) (bool, error)
	// AddGroupingPolicy 用户绑定角色
	AddGroupingPolicy(ctx context.Context, req domain.AddGroupingPolicy) (bool, error)
	// CreateOrUpdateFilteredPolicies 通过过滤完成新增修改删除
	CreateOrUpdateFilteredPolicies(ctx context.Context, req domain.Policies) (bool, error)
	// Authorize 权限校验
	Authorize(ctx context.Context, userId, path, method string) (bool, error)
	// GetImplicitPermissionsForUser 获取用户拥有的所有接口权限
	GetImplicitPermissionsForUser(ctx context.Context, userId int64) ([]domain.Policy, error)
	// GetPermissionsForRole 获取角色拥有的所有权限
	GetPermissionsForRole(ctx context.Context, roleCode string) ([]domain.Policy, error)
	// UpdateFilteredGrouping 考虑到一个用户不会拥有太多角色，先删除、在重新添加用户角色
	UpdateFilteredGrouping(ctx context.Context, userId int64, roleCodes []string) (bool, error)
	// GetRolesForUser 获取用户的所有角色
	GetRolesForUser(ctx context.Context, userId int64) ([]string, error)
}

type service struct {
	enforcer *casbin.SyncedEnforcer
}

func (s *service) GetRolesForUser(ctx context.Context, userId int64) ([]string, error) {
	return s.enforcer.GetRolesForUser(strconv.FormatInt(userId, 10))
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

func (s *service) GetImplicitPermissionsForUser(ctx context.Context, userId int64) ([]domain.Policy, error) {
	pers, err := s.enforcer.GetImplicitPermissionsForUser(strconv.FormatInt(userId, 10))
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
	return s.enforcer.Enforce(userId, path, method)
}

func (s *service) AddPolicies(ctx context.Context, req domain.Policies) (bool, error) {
	var policies [][]string
	for _, policy := range req.Policies {
		policies = append(policies, []string{req.RoleCode, policy.Path, policy.Method, policy.Effect.ToString()})
	}

	ok, err := s.enforcer.RemovePolicies(policies)
	if err != nil {
		return ok, err
	}

	ok, err = s.enforcer.AddPolicies(policies)
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
