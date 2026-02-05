package service

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/Duke1616/ecmdb/internal/permission/internal/domain"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/role"
	"github.com/Duke1616/ecmdb/pkg/tools"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	// MenuChangeTriggerRoleAndPolicy 菜单绑定 API 发生变化，进行权限同步
	MenuChangeTriggerRoleAndPolicy(ctx context.Context, action uint8, req domain.Menu) error

	// AddPermissionForRole 对指定的角色添加菜单权限
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

	// 提取所有 endpoints
	var allEndpoints []menu.Endpoint
	for _, m := range menus {
		allEndpoints = append(allEndpoints, m.Endpoints...)
	}

	// 转换为 Policy
	policies := slice.Map(allEndpoints, func(idx int, src menu.Endpoint) policy.Policy {
		return policy.Policy{
			Path:     src.Path,
			Method:   src.Method,
			Resource: src.Resource,
			Effect:   "allow",
		}
	})

	// 去重
	policies = tools.UniqueBy(policies, func(p policy.Policy) string {
		return fmt.Sprintf("%s:%s:%s", p.Method, p.Path, p.Resource)
	})

	// 同步权限
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
		// 新增菜单，自动授权给 admin 超级管理员
		if err = s.initialAdminPermission(ctx, req); err != nil {
			return err
		}
		s.logger.Info("菜单权限同步完成: 新增默认Admin权限", elog.Int64("menuId", req.Id))
	case domain.WRITE.ToUint8():
		var changed bool
		changed, err = s.write(ctx, roles, req.Endpoints)
		if err != nil {
			return err
		}
		s.logger.Info("菜单权限同步完成: 更新权限", elog.Int64("menuId", req.Id), elog.Any("changed", changed))
	case domain.DELETE.ToUint8():
		var changed bool
		changed, err = s.remove(ctx, roles, req.Endpoints)
		if err != nil {
			return err
		}
		s.logger.Info("菜单权限同步完成: 删除权限", elog.Int64("menuId", req.Id), elog.Any("changed", changed))
	}
	return nil
}

func (s *service) initialAdminPermission(ctx context.Context, req domain.Menu) error {
	var eg errgroup.Group
	if len(req.Endpoints) != 0 {
		eg.Go(func() error {
			_, err := s.write(ctx, []role.Role{{Code: role.AdminRole}}, req.Endpoints)
			return err
		})
	}

	eg.Go(func() error {
		// 获取角色
		r, err := s.roleSvc.FindByRoleCode(ctx, role.AdminRole)
		if err != nil {
			return err
		}

		// 组合菜单ID, 进行更新
		menus := append(r.MenuIds, req.Id)
		_, err = s.roleSvc.CreateOrUpdateRoleMenuIds(ctx, role.AdminRole, menus)
		return err
	})

	return eg.Wait()
}

func (s *service) write(ctx context.Context, roles []role.Role, es []domain.Endpoint) (bool, error) {
	return s.policySvc.AddBatchPolicies(ctx, s.toBatchPolicies(roles, es))
}

func (s *service) remove(ctx context.Context, roles []role.Role, es []domain.Endpoint) (bool, error) {
	return s.policySvc.RemoveBatchPolicies(ctx, s.toBatchPolicies(roles, es))
}

func (s *service) toBatchPolicies(roles []role.Role, es []domain.Endpoint) policy.BatchPolicies {
	policies := slice.Map(roles, func(idx int, r role.Role) policy.Policies {
		return policy.Policies{
			RoleCode: r.Code,
			Policies: slice.Map(es, func(idx int, src domain.Endpoint) policy.Policy {
				return policy.Policy{
					Path:     src.Path,
					Method:   src.Method,
					Resource: src.Resource,
					Effect:   "allow",
				}
			}),
		}
	})

	return policy.BatchPolicies{Policies: policies}
}

func NewService(roleSvc role.Service, policySvc policy.Service, menuSvc menu.Service) Service {
	return &service{
		logger:    elog.DefaultLogger.With(elog.FieldComponentName("PermissionService")),
		roleSvc:   roleSvc,
		policySvc: policySvc,
		menuSvc:   menuSvc,
	}
}
