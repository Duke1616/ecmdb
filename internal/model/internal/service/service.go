package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository"
)

type Service interface {
	CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error)
}

type service struct {
	repo repository.ModelRepository
}

func NewService(repo repository.ModelRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error) {
	return s.repo.CreateModelGroup(ctx, req)
}
