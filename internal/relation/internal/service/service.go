package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
)

type Service interface {
	CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error)
	CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error)
	ListModelRelation(ctx context.Context, offset, limit int64) ([]domain.ModelRelation, int64, error)
	ListResourceRelation(ctx context.Context, offset, limit int64) ([]domain.ResourceRelation, int64, error)
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

func (s *service) ListModelRelation(ctx context.Context, offset, limit int64) ([]domain.ModelRelation, int64, error) {
	relation, err := s.repo.ListModelRelation(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return relation, 0, nil
}

func (s *service) ListResourceRelation(ctx context.Context, offset, limit int64) ([]domain.ResourceRelation, int64, error) {
	relation, err := s.repo.ListResourceRelation(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return relation, 0, nil
}
