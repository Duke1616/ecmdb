package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/runner/internal/domain"
	"github.com/Duke1616/ecmdb/internal/runner/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	Register(ctx context.Context, req domain.Runner) (int64, error)
	ListRunner(ctx context.Context, offset, limit int64) ([]domain.Runner, int64, error)
}

type service struct {
	repo repository.RunnerRepository
}

func NewService(repo repository.RunnerRepository) Service {
	return &service{
		repo: repo,
	}
}
func (s *service) Register(ctx context.Context, req domain.Runner) (int64, error) {
	return s.repo.RegisterRunner(ctx, req)
}

func (s *service) ListRunner(ctx context.Context, offset, limit int64) ([]domain.Runner, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Runner
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListRunner(ctx, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}
