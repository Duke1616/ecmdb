package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/department/internal/domain"
	"github.com/Duke1616/ecmdb/internal/department/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type DepartmentRepository interface {
	CreateDepartment(ctx context.Context, req domain.Department) (int64, error)
	UpdateDepartment(ctx context.Context, req domain.Department) (int64, error)
	ListDepartment(ctx context.Context) ([]domain.Department, error)
	Total(ctx context.Context) (int64, error)
}

func NewDepartmentRepository(dao dao.DepartmentDAO) DepartmentRepository {
	return &departmentRepository{
		dao: dao,
	}
}

type departmentRepository struct {
	dao dao.DepartmentDAO
}

func (repo *departmentRepository) UpdateDepartment(ctx context.Context, req domain.Department) (int64, error) {
	return repo.dao.UpdateDepartment(ctx, repo.toEntity(req))
}

func (repo *departmentRepository) ListDepartment(ctx context.Context) ([]domain.Department, error) {
	departments, err := repo.dao.ListDepartment(ctx)
	return slice.Map(departments, func(idx int, src dao.Department) domain.Department {
		return repo.toDomain(src)
	}), err
}

func (repo *departmentRepository) Total(ctx context.Context) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (repo *departmentRepository) CreateDepartment(ctx context.Context, req domain.Department) (int64, error) {
	return repo.dao.CreateDepartment(ctx, repo.toEntity(req))
}

func (repo *departmentRepository) toEntity(req domain.Department) dao.Department {
	return dao.Department{
		Id:         req.Id,
		Pid:        req.Pid,
		Name:       req.Name,
		Sort:       req.Sort,
		Enabled:    req.Enabled,
		Leaders:    req.Leaders,
		MainLeader: req.MainLeader,
	}
}

func (repo *departmentRepository) toDomain(req dao.Department) domain.Department {
	return domain.Department{
		Id:         req.Id,
		Pid:        req.Pid,
		Name:       req.Name,
		Sort:       req.Sort,
		Enabled:    req.Enabled,
		Leaders:    req.Leaders,
		MainLeader: req.MainLeader,
	}
}
