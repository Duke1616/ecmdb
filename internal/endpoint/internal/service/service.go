package service

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/endpoint/internal/domain"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	// RegisterEndpoint 注册单个端点
	RegisterEndpoint(ctx context.Context, req domain.Endpoint) (int64, error)

	// BatchRegisterByResource 按 Resource 批量注册端点
	// 支持智能同步：插入新端点、更新已存在端点、删除不再存在的端点
	BatchRegisterByResource(ctx context.Context, resource string, req []domain.Endpoint) (int64, error)

	// ListEndpoints 获取列表
	ListEndpoints(ctx context.Context, offset, limit int64, path string) ([]domain.Endpoint, int64, error)
}

type service struct {
	repo repository.EndpointRepository
}

func NewService(repo repository.EndpointRepository) Service {
	return &service{
		repo: repo,
	}
}

// BatchRegisterByResource 按 Resource 批量注册端点
func (s *service) BatchRegisterByResource(ctx context.Context, resource string, req []domain.Endpoint) (int64, error) {
	if len(req) == 0 {
		return 0, nil
	}

	// 直接调用 repository 的按 Resource 注册方法
	return s.repo.BatchRegisterByResource(ctx, resource, req)
}

func (s *service) ListEndpoints(ctx context.Context, offset, limit int64, path string) ([]domain.Endpoint, int64, error) {
	var (
		eg    errgroup.Group
		es    []domain.Endpoint
		total int64
	)
	eg.Go(func() error {
		var err error
		es, err = s.repo.ListEndpoint(ctx, offset, limit, path)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx, path)
		return err
	})
	if err := eg.Wait(); err != nil {
		return es, total, err
	}
	return es, total, nil
}

func (s *service) RegisterEndpoint(ctx context.Context, req domain.Endpoint) (int64, error) {
	return s.repo.CreateEndpoint(ctx, req)
}
