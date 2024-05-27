package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	CreateResource(ctx context.Context, req domain.Resource) (int64, error)
	FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error)
	ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]domain.Resource, int64, error)

	// ListResourceByIds 资源关联关系调用，查询关联数据
	ListResourceByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error)

	// ListExcludeResourceByIds 排除部分的 ids
	ListExcludeResourceByIds(ctx context.Context, fields []string, modelUid string, offset, limit int64, ids []int64) ([]domain.Resource, int64, error)

	DeleteResource(ctx context.Context, id int64) (int64, error)

	// PipelineByModelUid 聚合查看模型下的数量
	PipelineByModelUid(ctx context.Context) (map[string]int, error)

	Search(ctx context.Context, text string) ([]domain.SearchResource, error)
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

func (s *service) ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]domain.Resource, int64, error) {
	var (
		total     int64
		resources []domain.Resource
		eg        errgroup.Group
	)
	eg.Go(func() error {
		var err error
		resources, err = s.repo.ListResource(ctx, fields, modelUid, offset, limit)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx, modelUid)
		return err
	})
	if err := eg.Wait(); err != nil {
		return resources, total, err
	}
	return resources, total, nil
}

func (s *service) ListResourceByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error) {
	return s.repo.ListResourcesByIds(ctx, fields, ids)
}

func (s *service) ListExcludeResourceByIds(ctx context.Context, fields []string, modelUid string, offset, limit int64, ids []int64) ([]domain.Resource, int64, error) {
	var (
		total     int64
		resources []domain.Resource
		eg        errgroup.Group
	)
	eg.Go(func() error {
		var err error
		resources, err = s.repo.ListExcludeResourceByIds(ctx, fields, modelUid, offset, limit, ids)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalExcludeResourceByIds(ctx, modelUid, ids)
		return err
	})
	if err := eg.Wait(); err != nil {
		return resources, total, err
	}
	return resources, total, nil
}

func (s *service) DeleteResource(ctx context.Context, id int64) (int64, error) {
	return s.repo.DeleteResource(ctx, id)
}

func (s *service) PipelineByModelUid(ctx context.Context) (map[string]int, error) {
	return s.repo.PipelineByModelUid(ctx)
}

func (s *service) Search(ctx context.Context, text string) ([]domain.SearchResource, error) {
	return s.repo.Search(ctx, text)
}
