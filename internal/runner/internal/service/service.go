package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/runner/internal/domain"
	"github.com/Duke1616/ecmdb/internal/runner/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	Register(ctx context.Context, req domain.Runner) (int64, error)
	Update(ctx context.Context, req domain.Runner) (int64, error)
	Detail(ctx context.Context, id int64) (domain.Runner, error)
	Delete(ctx context.Context, id int64) (int64, error)
	ListRunner(ctx context.Context, offset, limit int64) ([]domain.Runner, int64, error)
	FindByCodebookUid(ctx context.Context, codebookUid string, tag string) (domain.Runner, error)
	ListTagsPipelineByCodebookUid(ctx context.Context) ([]domain.RunnerTags, error)
}

type service struct {
	repo repository.RunnerRepository
}

func (s *service) Detail(ctx context.Context, id int64) (domain.Runner, error) {
	return s.repo.Detail(ctx, id)
}

func (s *service) Delete(ctx context.Context, id int64) (int64, error) {
	return s.repo.Delete(ctx, id)
}

func (s *service) Update(ctx context.Context, req domain.Runner) (int64, error) {
	return s.repo.Update(ctx, req)
}

func (s *service) ListTagsPipelineByCodebookUid(ctx context.Context) ([]domain.RunnerTags, error) {
	return s.repo.ListTagsPipelineByCodebookUid(ctx)
}

func (s *service) FindByCodebookUid(ctx context.Context, codebookUid string, tag string) (domain.Runner, error) {
	return s.repo.FindByCodebookUid(ctx, codebookUid, tag)
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
