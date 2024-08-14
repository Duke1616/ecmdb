package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/domain"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	RegisterEndpoint(ctx context.Context, req domain.Endpoint) (int64, error)
	RegisterMutilEndpoint(ctx context.Context, req []domain.Endpoint) (int64, error)
	ListResource(ctx context.Context, offset, limit int64, path string) ([]domain.Endpoint, int64, error)
}

type service struct {
	repo repository.EndpointRepository
}

func NewService(repo repository.EndpointRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) RegisterMutilEndpoint(ctx context.Context, req []domain.Endpoint) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) ListResource(ctx context.Context, offset, limit int64, path string) ([]domain.Endpoint, int64, error) {
	var (
		eg    errgroup.Group
		es    []domain.Endpoint
		total int64
	)
	eg.Go(func() error {
		var err error
		es, err = s.repo.ListEndpoint(ctx, offset, limit, path)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx, path)
		return err
	})
	if err := eg.Wait(); err != nil {
		return es, total, err
	}
	return es, total, nil
}

func (s *service) RegisterEndpoint(ctx context.Context, req domain.Endpoint) (int64, error) {
	return s.repo.CreateEndpoint(ctx, req)
}
