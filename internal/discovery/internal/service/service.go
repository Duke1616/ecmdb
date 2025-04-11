package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/discovery/internal/domain"
	"github.com/Duke1616/ecmdb/internal/discovery/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	Create(ctx context.Context, req domain.Discovery) (int64, error)
	Update(ctx context.Context, req domain.Discovery) (int64, error)
	Delete(ctx context.Context, id int64) (int64, error)
	ListByTemplateId(ctx context.Context, offset, limit int64, templateId int64) ([]domain.Discovery, int64, error)
}

type service struct {
	repo repository.DiscoveryRepository
}

func (s *service) Delete(ctx context.Context, id int64) (int64, error) {
	return s.repo.Delete(ctx, id)
}

func (s *service) Create(ctx context.Context, req domain.Discovery) (int64, error) {
	return s.repo.Create(ctx, req)
}

func (s *service) Update(ctx context.Context, req domain.Discovery) (int64, error) {
	return s.repo.Update(ctx, req)
}

func (s *service) ListByTemplateId(ctx context.Context, offset, limit int64, templateId int64) ([]domain.Discovery, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Discovery
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListByTemplateId(ctx, offset, limit, templateId)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountByTemplateId(ctx, templateId)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func NewService(repo repository.DiscoveryRepository) Service {
	return &service{
		repo: repo,
	}
}
