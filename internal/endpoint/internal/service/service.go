package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/domain"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/repository"
)

type Service interface {
	RegisterEndpoint(ctx context.Context, req domain.Endpoint) (int64, error)
}

type service struct {
	repo repository.EndpointRepository
}

func NewService(repo repository.EndpointRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) RegisterEndpoint(ctx context.Context, req domain.Endpoint) (int64, error) {
	return s.repo.CreateEndpoint(ctx, req)
}
