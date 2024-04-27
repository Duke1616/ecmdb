package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository"
)

type Service interface {
	CreateResource(ctx context.Context, req domain.Resource) (int64, error)
	FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error)
	ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]domain.Resource, error)

	// ListResourceByIds 资源关联关系调用，查询关联数据
	ListResourceByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error)

	// ListExcludeResourceByids 排除部分的 ids
	ListExcludeResourceByids(ctx context.Context, fields []string, modelUid string, offset, limit int64, ids []int64) ([]domain.Resource, error)
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

func (s *service) FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error) {
	return s.repo.FindResourceById(ctx, fields, id)
}

func (s *service) ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]domain.Resource, error) {
	return s.repo.ListResource(ctx, fields, modelUid, offset, limit)
}

func (s *service) ListResourceByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error) {
	return s.repo.ListResourcesByIds(ctx, fields, ids)
}

func (s *service) ListExcludeResourceByids(ctx context.Context, fields []string, modelUid string, offset, limit int64, ids []int64) ([]domain.Resource, error) {
	return s.repo.ListExcludeResourceByids(ctx, fields, modelUid, offset, limit, ids)
}
