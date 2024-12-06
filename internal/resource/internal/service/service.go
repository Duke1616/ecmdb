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
	ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]domain.Resource,
		int64, error)
	CountByModelUid(ctx context.Context, modelUid string) (int64, error)
	SetCustomField(ctx context.Context, id int64, field string, data interface{}) (int64, error)
	// ListResourceByIds 资源关联关系调用，查询关联数据
	ListResourceByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error)
	// ListExcludeAndFilterResourceByIds 排序以及过滤
	ListExcludeAndFilterResourceByIds(ctx context.Context, fields []string, modelUid string, offset, limit int64,
		ids []int64, filter domain.Condition) ([]domain.Resource, int64, error)
	DeleteResource(ctx context.Context, id int64) (int64, error)
	// CountByModelUids 聚合查看模型下的数量
	CountByModelUids(ctx context.Context, modelUids []string) (map[string]int, error)
	Search(ctx context.Context, text string) ([]domain.SearchResource, error)

	FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error)
	UpdateResource(ctx context.Context, resource domain.Resource) (int64, error)
}

type service struct {
	repo repository.ResourceRepository
}

func (s *service) SetCustomField(ctx context.Context, id int64, field string, data interface{}) (int64, error) {
	return s.repo.SetCustomField(ctx, id, field, data)
}

func (s *service) CountByModelUid(ctx context.Context, modelUid string) (int64, error) {
	return s.repo.TotalByModelUid(ctx, modelUid)
}

func NewService(repo repository.ResourceRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) UpdateResource(ctx context.Context, resource domain.Resource) (int64, error) {
	return s.repo.UpdateResource(ctx, resource)
}

func (s *service) CreateResource(ctx context.Context, req domain.Resource) (int64, error) {
	// TODO 判断是否为加密模型，
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
		total, err = s.repo.TotalByModelUid(ctx, modelUid)
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

func (s *service) ListExcludeAndFilterResourceByIds(ctx context.Context, fields []string, modelUid string, offset,
	limit int64, ids []int64, filter domain.Condition) ([]domain.Resource, int64, error) {
	var (
		total     int64
		resources []domain.Resource
		eg        errgroup.Group
	)
	eg.Go(func() error {
		var err error
		resources, err = s.repo.ListExcludeAndFilterResourceByIds(ctx, fields, modelUid, offset, limit, ids, filter)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalExcludeAndFilterResourceByIds(ctx, modelUid, ids, filter)
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

func (s *service) CountByModelUids(ctx context.Context, modelUids []string) (map[string]int, error) {
	return s.repo.CountByModelUids(ctx, modelUids)
}

func (s *service) Search(ctx context.Context, text string) ([]domain.SearchResource, error) {
	return s.repo.Search(ctx, text)
}

func (s *service) FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error) {
	return s.repo.FindSecureData(ctx, id, fieldUid)
}

func (s *service) desensitization(ctx context.Context, modelUid string) ([]domain.Resource, error) {
	return nil, nil
}
