package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
)

type Service interface {
	CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error)
	CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error)
}

type service struct {
	repo repository.RelationRepository
}

func NewService(repo repository.RelationRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error) {
	return s.repo.CreateModelRelation(ctx, req)
}

func (s *service) CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error) {
	return s.repo.CreateResourceRelation(ctx, req)
}
