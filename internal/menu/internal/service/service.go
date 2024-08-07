package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/menu/internal/domain"
	"github.com/Duke1616/ecmdb/internal/menu/internal/repository"
)

type Service interface {
	CreateMenu(ctx context.Context, req domain.Menu) (int64, error)
}

type service struct {
	repo repository.MenuRepository
}

func (s *service) CreateMenu(ctx context.Context, req domain.Menu) (int64, error) {
	return s.repo.CreateMenu(ctx, req)
}

func NewService(repo repository.MenuRepository) Service {
	return &service{
		repo: repo,
	}
}
