package service

import (
	"context"
	"errors"
	"github.com/Duke1616/ecmdb/internal/codebook/internal/domain"
	"github.com/Duke1616/ecmdb/internal/codebook/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	CreateCodebook(ctx context.Context, req domain.Codebook) (int64, error)
	DetailCodebook(ctx context.Context, id int64) (domain.Codebook, error)
	ListCodebook(ctx context.Context, offset, limit int64) ([]domain.Codebook, int64, error)
	UpdateCodebook(ctx context.Context, req domain.Codebook) (int64, error)
	DeleteCodebook(ctx context.Context, id int64) (int64, error)
	ValidationSecret(ctx context.Context, identifier string, secret string) (bool, error)
	FindByUid(ctx context.Context, identifier string) (domain.Codebook, error)
}

type service struct {
	repo repository.CodebookRepository
}

func (s *service) FindByUid(ctx context.Context, identifier string) (domain.Codebook, error) {
	return s.repo.FindByUid(ctx, identifier)
}

func (s *service) ValidationSecret(ctx context.Context, identifier string, secret string) (bool, error) {
	_, err := s.repo.FindBySecret(ctx, identifier, secret)
	if !errors.Is(err, repository.ErrUserNotFound) {
		return true, err
	}

	return false, err
}

func NewService(repo repository.CodebookRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateCodebook(ctx context.Context, req domain.Codebook) (int64, error) {
	return s.repo.CreateCodebook(ctx, req)
}

func (s *service) DetailCodebook(ctx context.Context, id int64) (domain.Codebook, error) {
	return s.repo.DetailCodebook(ctx, id)
}

func (s *service) ListCodebook(ctx context.Context, offset, limit int64) ([]domain.Codebook, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Codebook
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListCodebook(ctx, offset, limit)
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

func (s *service) UpdateCodebook(ctx context.Context, req domain.Codebook) (int64, error) {
	return s.repo.UpdateCodebook(ctx, req)
}

func (s *service) DeleteCodebook(ctx context.Context, id int64) (int64, error) {
	return s.repo.DeleteCodebook(ctx, id)
}
