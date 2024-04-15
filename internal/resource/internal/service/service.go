package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository"
	"github.com/Duke1616/ecmdb/pkg/mongox"
)

type Service interface {
	CreateResource(ctx context.Context, req domain.Resource) (int64, error)
	FindResourceById(ctx context.Context, req domain.DetailResource) ([]mongox.MapStr, error)
}

type service struct {
	repo repository.ResourceRepository
}

func NewService(repo repository.ResourceRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateResource(ctx context.Context, req domain.Resource) (int64, error) {
	return s.repo.CreateResource(ctx, req)
}

func (s *service) FindResourceById(ctx context.Context, req domain.DetailResource) ([]mongox.MapStr, error) {
	return s.repo.FindResourceById(ctx, req)
}

func (s *service) ListRelations(ctx context.Context, modelUIds []string) ([]domain.ResourceRelation, error) {
	return nil, nil
}
