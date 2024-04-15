package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error)
	CreateModel(ctx context.Context, req domain.Model) (int64, error)
	FindModelByUid(ctx context.Context, Identifies string) (domain.Model, error)
	ListModels(ctx context.Context, offset, limit int64) ([]domain.Model, int64, error)
}

type service struct {
	repo repository.ModelRepository
}

func NewService(repo repository.ModelRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateModelGroup(ctx context.Context, req domain.ModelGroup) (int64, error) {
	return s.repo.CreateModelGroup(ctx, req)
}

func (s *service) CreateModel(ctx context.Context, req domain.Model) (int64, error) {
	return s.repo.CreateModel(ctx, req)
}

func (s *service) FindModelByUid(ctx context.Context, Identifies string) (domain.Model, error) {
	return s.repo.FindModelByUid(ctx, Identifies)
}

func (s *service) ListModels(ctx context.Context, offset, limit int64) ([]domain.Model, int64, error) {
	var (
		total     int64
		modelList []domain.Model
		eg        errgroup.Group
	)
	eg.Go(func() error {
		var err error
		modelList, err = s.repo.ListModels(ctx, offset, limit)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx)
		return err
	})
	if err := eg.Wait(); err != nil {
		return modelList, total, err
	}
	return modelList, total, nil
}
