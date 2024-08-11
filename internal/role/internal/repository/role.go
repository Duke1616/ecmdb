package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/role/internal/domain"
	"github.com/Duke1616/ecmdb/internal/role/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type RoleRepository interface {
	CreateRole(ctx context.Context, req domain.Role) (int64, error)
	ListRole(ctx context.Context, offset, limit int64) ([]domain.Role, error)
	Total(ctx context.Context) (int64, error)
	UpdateRole(ctx context.Context, req domain.Role) (int64, error)
	FindByIncludeCodes(ctx context.Context, codes []string) ([]domain.Role, error)
	FindByExcludeCodes(ctx context.Context, offset, limit int64, codes []string) ([]domain.Role, error)
	CountByExcludeCodes(ctx context.Context, codes []string) (int64, error)
}

type roleRepository struct {
	dao dao.RoleDAO
}

func (repo *roleRepository) FindByIncludeCodes(ctx context.Context, codes []string) ([]domain.Role, error) {
	rs, err := repo.dao.FindByIncludeCodes(ctx, codes)
	return slice.Map(rs, func(idx int, src dao.Role) domain.Role {
		return repo.toDomain(src)
	}), err
}

func (repo *roleRepository) FindByExcludeCodes(ctx context.Context, offset, limit int64, codes []string) ([]domain.Role, error) {
	rs, err := repo.dao.FindByExcludeCodes(ctx, offset, limit, codes)
	return slice.Map(rs, func(idx int, src dao.Role) domain.Role {
		return repo.toDomain(src)
	}), err
}

func (repo *roleRepository) CountByExcludeCodes(ctx context.Context, codes []string) (int64, error) {
	return repo.dao.CountByExcludeCodes(ctx, codes)
}

func (repo *roleRepository) UpdateRole(ctx context.Context, req domain.Role) (int64, error) {
	return repo.dao.UpdateRole(ctx, repo.toEntity(req))
}

func (repo *roleRepository) ListRole(ctx context.Context, offset, limit int64) ([]domain.Role, error) {
	rs, err := repo.dao.ListRole(ctx, offset, limit)
	return slice.Map(rs, func(idx int, src dao.Role) domain.Role {
		return repo.toDomain(src)
	}), err
}

func (repo *roleRepository) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func (repo *roleRepository) CreateRole(ctx context.Context, req domain.Role) (int64, error) {
	return repo.dao.CreateRole(ctx, repo.toEntity(req))
}

func NewRoleRepository(dao dao.RoleDAO) RoleRepository {
	return &roleRepository{
		dao: dao,
	}
}

func (repo *roleRepository) toEntity(req domain.Role) dao.Role {
	return dao.Role{
		Id:     req.Id,
		Name:   req.Name,
		Code:   req.Code,
		Desc:   req.Desc,
		Status: req.Status,
	}
}

func (repo *roleRepository) toDomain(req dao.Role) domain.Role {
	return domain.Role{
		Id:     req.Id,
		Name:   req.Name,
		Code:   req.Code,
		Desc:   req.Desc,
		Status: req.Status,
	}
}
