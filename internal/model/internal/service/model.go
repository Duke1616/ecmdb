package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	Create(ctx context.Context, req domain.Model) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]domain.Model, int64, error)
	ListAll(ctx context.Context) ([]domain.Model, error)

	DeleteById(ctx context.Context, id int64) (int64, error)
	DeleteByModelUid(ctx context.Context, modelUid string) (int64, error)
	FindModelById(ctx context.Context, id int64) (domain.Model, error)
	ListModelByGroupIds(ctx context.Context, mgids []int64) ([]domain.Model, error)
}

type service struct {
	repo repository.ModelRepository
}

func NewModelService(repo repository.ModelRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) ListAll(ctx context.Context) ([]domain.Model, error) {
	return s.repo.ListAll(ctx)
}

func (s *service) Create(ctx context.Context, req domain.Model) (int64, error) {
	return s.repo.Create(ctx, req)
}

func (s *service) FindModelById(ctx context.Context, id int64) (domain.Model, error) {
	return s.repo.FindById(ctx, id)
}

func (s *service) List(ctx context.Context, offset, limit int64) ([]domain.Model, int64, error) {
	var (
		total  int64
		models []domain.Model
		eg     errgroup.Group
	)
	eg.Go(func() error {
		var err error
		models, err = s.repo.List(ctx, offset, limit)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx)
		return err
	})
	if err := eg.Wait(); err != nil {
		return models, total, err
	}
	return models, total, nil
}

func (s *service) ListModelByGroupIds(ctx context.Context, mgids []int64) ([]domain.Model, error) {
	return s.repo.ListByGroupIds(ctx, mgids)
}

func (s *service) DeleteById(ctx context.Context, id int64) (int64, error) {
	return s.repo.DeleteById(ctx, id)
}

func (s *service) DeleteByModelUid(ctx context.Context, modelUid string) (int64, error) {
	return s.repo.DeleteByUid(ctx, modelUid)
}
