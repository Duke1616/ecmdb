package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/role/internal/domain"
	"github.com/Duke1616/ecmdb/internal/role/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	// CreateRole 创建角色
	CreateRole(ctx context.Context, req domain.Role) (int64, error)
	// ListRole 获取角色列表
	ListRole(ctx context.Context, offset, limit int64) ([]domain.Role, int64, error)
	// UpdateRole 变更角色信息
	UpdateRole(ctx context.Context, req domain.Role) (int64, error)
	// CreateOrUpdateRoleMenuIds 新增角色的菜单权限
	CreateOrUpdateRoleMenuIds(ctx context.Context, code string, menuIds []int64) (int64, error)
	// FindByExcludeCodes 查找排除当前角色编码的数据
	FindByExcludeCodes(ctx context.Context, offset, limit int64, codes []string) ([]domain.Role, int64, error)
	// FindByIncludeCodes 查找包含当前角色编码的数据
	FindByIncludeCodes(ctx context.Context, codes []string) ([]domain.Role, error)
	// FindByMenuId 查找包含菜单 ID 的角色
	FindByMenuId(ctx context.Context, menuId int64) ([]domain.Role, error)
	// FindByRoleCode 查找角色编码数据
	FindByRoleCode(ctx context.Context, code string) (domain.Role, error)
	// DeleteRole 删除角色
	DeleteRole(ctx context.Context, id int64) (int64, error)
}

type service struct {
	repo repository.RoleRepository
}

func (s *service) DeleteRole(ctx context.Context, id int64) (int64, error) {
	return s.repo.DeleteRole(ctx, id)
}

func (s *service) FindByRoleCode(ctx context.Context, code string) (domain.Role, error) {
	return s.repo.FindByRoleCode(ctx, code)
}

func (s *service) FindByMenuId(ctx context.Context, menuId int64) ([]domain.Role, error) {
	return s.repo.FindByMenuId(ctx, menuId)
}

func (s *service) CreateOrUpdateRoleMenuIds(ctx context.Context, code string, menuIds []int64) (int64, error) {
	return s.repo.CreateOrUpdateRoleMenuIds(ctx, code, menuIds)
}

func (s *service) FindByExcludeCodes(ctx context.Context, offset, limit int64, codes []string) ([]domain.Role, int64, error) {
	var (
		eg    errgroup.Group
		rs    []domain.Role
		total int64
	)
	eg.Go(func() error {
		var err error
		rs, err = s.repo.FindByExcludeCodes(ctx, offset, limit, codes)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountByExcludeCodes(ctx, codes)
		return err
	})
	if err := eg.Wait(); err != nil {
		return rs, total, err
	}
	return rs, total, nil
}

func (s *service) FindByIncludeCodes(ctx context.Context, codes []string) ([]domain.Role, error) {
	return s.repo.FindByIncludeCodes(ctx, codes)
}

func (s *service) UpdateRole(ctx context.Context, req domain.Role) (int64, error) {
	return s.repo.UpdateRole(ctx, req)
}

func (s *service) ListRole(ctx context.Context, offset, limit int64) ([]domain.Role, int64, error) {
	var (
		eg    errgroup.Group
		rs    []domain.Role
		total int64
	)
	eg.Go(func() error {
		var err error
		rs, err = s.repo.ListRole(ctx, offset, limit)
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

func (s *service) CreateRole(ctx context.Context, req domain.Role) (int64, error) {
	return s.repo.CreateRole(ctx, req)
}

func NewService(repo repository.RoleRepository) Service {
	return &service{
		repo: repo,
	}
}
