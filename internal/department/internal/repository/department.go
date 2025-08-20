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
	FindById(ctx context.Context, id int64) (domain.Department, error)
	DeleteDepartment(ctx context.Context, id int64) (int64, error)
	ListDepartment(ctx context.Context) ([]domain.Department, error)
	ListDepartmentByIds(ctx context.Context, ids []int64) ([]domain.Department, error)
}

func NewDepartmentRepository(dao dao.DepartmentDAO) DepartmentRepository {
	return &departmentRepository{
		dao: dao,
	}
}

type departmentRepository struct {
	dao dao.DepartmentDAO
}

func (repo *departmentRepository) FindById(ctx context.Context, id int64) (domain.Department, error) {
	department, err := repo.dao.FindByid(ctx, id)
	return repo.toDomain(department), err
}

func (repo *departmentRepository) ListDepartmentByIds(ctx context.Context, ids []int64) ([]domain.Department, error) {
	departments, err := repo.dao.ListDepartmentByIds(ctx, ids)
	return slice.Map(departments, func(idx int, src dao.Department) domain.Department {
		return repo.toDomain(src)
	}), err
}

func (repo *departmentRepository) DeleteDepartment(ctx context.Context, id int64) (int64, error) {
	return repo.dao.DeleteDepartment(ctx, id)
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
