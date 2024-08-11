package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/role/internal/domain"
	"github.com/Duke1616/ecmdb/internal/role/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	CreateRole(ctx context.Context, req domain.Role) (int64, error)
	ListRole(ctx context.Context, offset, limit int64) ([]domain.Role, int64, error)
	UpdateRole(ctx context.Context, req domain.Role) (int64, error)
	FindByExcludeCodes(ctx context.Context, offset, limit int64, codes []string) ([]domain.Role, int64, error)
	FindByIncludeCodes(ctx context.Context, codes []string) ([]domain.Role, error)
}

type service struct {
	repo repository.RoleRepository
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
