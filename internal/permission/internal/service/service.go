package service

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/Duke1616/ecmdb/internal/permission/internal/domain"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/role"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

type Service interface {
	MenuChangeTriggerRoleAndPolicy(ctx context.Context, action uint8, req domain.Menu) error
	AddPermissionForRole(ctx context.Context, RoleCode string, menuIds []int64) error
}

type service struct {
	logger    *elog.Component
	roleSvc   role.Service
	menuSvc   menu.Service
	policySvc policy.Service
}

func (s *service) AddPermissionForRole(ctx context.Context, RoleCode string, menuIds []int64) error {
	// 查询需要添加权限的菜单信息
	menus, err := s.menuSvc.FindByIds(ctx, menuIds)
	if err != nil {
		return err
	}

	// 根据菜单信息，查询API接口权限，当前只做allow，不考虑deny的情况
	var policies []policy.Policy
	uniquePermissions := make(map[policy.Policy]bool)
	for _, m := range menus {
		ps := slice.FilterMap(m.Endpoints, func(idx int, src menu.Endpoint) (policy.Policy, bool) {
			p := policy.Policy{
				Path:   src.Path,
				Method: src.Method,
				Effect: "allow",
			}
			if !uniquePermissions[p] {
				uniquePermissions[p] = true
				return p, true
			}

			return policy.Policy{}, false
		})

		policies = append(policies, ps...)
	}

	_, err = s.policySvc.CreateOrUpdateFilteredPolicies(ctx, policy.Policies{
		RoleCode: RoleCode,
		Policies: policies,
	})

	return nil
}

func (s *service) MenuChangeTriggerRoleAndPolicy(ctx context.Context, action uint8, req domain.Menu) error {
	roles, err := s.roleSvc.FindByMenuId(ctx, req.Id)
	if err != nil {
		return fmt.Errorf("根据菜单ID: %d, 获取角色失败, %w", req.Id, err)
	}

	switch action {
	case domain.CREATE.ToUint8():
		// TODO 处理超级管理员 Admin 角色情况, 任何菜单新增都添加对应菜单、API权限
	case domain.WRITE.ToUint8():
		err = s.write(ctx, roles, req.Endpoints)
		if err != nil {
			return err
		}
	case domain.REWRITE.ToUint8():
		err = s.reWrite(ctx, roles)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *service) write(ctx context.Context, roles []role.Role, es []domain.Endpoint) error {
	for _, r := range roles {
		policies := slice.Map(es, func(idx int, src domain.Endpoint) policy.Policy {
			return policy.Policy{
				Path:   src.Path,
				Method: src.Method,
				Effect: "allow",
			}
		})

		_, err := s.policySvc.AddPolicies(ctx, policy.Policies{
			RoleCode: r.Code,
			Policies: policies,
		})

		if err != nil {
			return err
		}
	}
	return nil
}

func (s *service) reWrite(ctx context.Context, roles []role.Role) error {
	for _, r := range roles {
		err := s.AddPermissionForRole(ctx, r.Code, r.MenuIds)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewService(roleSvc role.Service, policySvc policy.Service, menuSvc menu.Service) Service {
	return &service{
		logger:    elog.DefaultLogger,
		roleSvc:   roleSvc,
		policySvc: policySvc,
		menuSvc:   menuSvc,
	}
}
