package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/rota/internal/domain"
	"github.com/Duke1616/ecmdb/internal/rota/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	Create(ctx context.Context, req domain.Rota) (int64, error)
	AddSchedulingRole(ctx context.Context, id int64, rr domain.RotaRule) (int64, error)
	UpdateSchedulingRole(ctx context.Context, id int64, rotaRules []domain.RotaRule) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]domain.Rota, int64, error)
	Detail(ctx context.Context, id int64) (domain.Rota, error)
}

func NewService(repo repository.RotaRepository) Service {
	return &service{
		repo: repo,
	}
}

type service struct {
	repo repository.RotaRepository
}

func (s *service) UpdateSchedulingRole(ctx context.Context, id int64, rotaRules []domain.RotaRule) (int64, error) {
	return s.repo.UpdateSchedulingRole(ctx, id, rotaRules)
}

func (s *service) Detail(ctx context.Context, id int64) (domain.Rota, error) {
	return s.repo.Detail(ctx, id)
}

func (s *service) List(ctx context.Context, offset, limit int64) ([]domain.Rota, int64, error) {
	var (
		eg    errgroup.Group
		rs    []domain.Rota
		total int64
	)
	eg.Go(func() error {
		var err error
		rs, err = s.repo.List(ctx, offset, limit)
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

func (s *service) Create(ctx context.Context, req domain.Rota) (int64, error) {
	return s.repo.Create(ctx, req)
}

func (s *service) AddSchedulingRole(ctx context.Context, id int64, rr domain.RotaRule) (int64, error) {
	return s.repo.AddSchedulingRole(ctx, id, rr)
}
