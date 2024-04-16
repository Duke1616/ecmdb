package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
)

type RelationResourceService interface {
	CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error)
	ListResourceRelation(ctx context.Context, offset, limit int64) ([]domain.ResourceRelation, int64, error)

	// ListResourceIds 通过关联类型和模型UID 获取关联的 resources ids
	ListResourceIds(ctx context.Context, modelUid string, relationType string) ([]int64, error)
}

type resourceService struct {
	repo repository.RelationResourceRepository
}

func NewRelationResourceService(repo repository.RelationResourceRepository) RelationResourceService {
	return &resourceService{
		repo: repo,
	}
}

func (s *resourceService) CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error) {
	return s.repo.CreateResourceRelation(ctx, req)
}

func (s *resourceService) ListResourceRelation(ctx context.Context, offset, limit int64) ([]domain.ResourceRelation, int64, error) {
	relation, err := s.repo.ListResourceRelation(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return relation, 0, nil
}

func (s *resourceService) ListResourceIds(ctx context.Context, modelUid string, relationType string) ([]int64, error) {
	return s.repo.ListResourceIds(ctx, modelUid, relationType)
}
