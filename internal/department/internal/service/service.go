package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/department/internal/domain"
	"github.com/Duke1616/ecmdb/internal/department/internal/repository"
)

type Service interface {
	CreateDepartment(ctx context.Context, req domain.Department) (int64, error)
	UpdateDepartment(ctx context.Context, req domain.Department) (int64, error)
	ListDepartment(ctx context.Context) ([]domain.Department, error)
}

type service struct {
	repo repository.DepartmentRepository
}

func (s *service) UpdateDepartment(ctx context.Context, req domain.Department) (int64, error) {
	return s.repo.UpdateDepartment(ctx, req)
}

func (s *service) ListDepartment(ctx context.Context) ([]domain.Department, error) {
	return s.repo.ListDepartment(ctx)
}

func (s *service) CreateDepartment(ctx context.Context, req domain.Department) (int64, error) {
	return s.repo.CreateDepartment(ctx, req)
}

func NewService(repo repository.DepartmentRepository) Service {
	return &service{
		repo: repo,
	}
}
