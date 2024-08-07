package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/menu/internal/domain"
	"github.com/Duke1616/ecmdb/internal/menu/internal/repository"
)

type Service interface {
	CreateMenu(ctx context.Context, req domain.Menu) (int64, error)
	UpdateMenu(ctx context.Context, req domain.Menu) (int64, error)
	ListMenu(ctx context.Context) ([]domain.Menu, error)
	GetAllMenu(ctx context.Context, userId string) ([]domain.Menu, error)
}

type service struct {
	repo repository.MenuRepository
}

func (s *service) UpdateMenu(ctx context.Context, req domain.Menu) (int64, error) {
	return s.repo.UpdateMenu(ctx, req)
}

func (s *service) ListMenu(ctx context.Context) ([]domain.Menu, error) {
	return s.repo.ListMenu(ctx)
}

func (s *service) GetAllMenu(ctx context.Context, userId string) ([]domain.Menu, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) CreateMenu(ctx context.Context, req domain.Menu) (int64, error) {
	return s.repo.CreateMenu(ctx, req)
}

func NewService(repo repository.MenuRepository) Service {
	return &service{
		repo: repo,
	}
}
