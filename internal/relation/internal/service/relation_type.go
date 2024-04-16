package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
)

type RelationTypeService interface {
	Create(ctx context.Context, req domain.RelationType) (int64, error)
}

type service struct {
	repo repository.RelationTypeRepository
}

func NewRelationTypeService(repo repository.RelationTypeRepository) RelationTypeService {
	return &service{
		repo: repo,
	}
}

func (s *service) Create(ctx context.Context, req domain.RelationType) (int64, error) {
	return s.repo.Create(ctx, req)
}
