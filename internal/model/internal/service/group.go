package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository"
	"golang.org/x/sync/errgroup"
)

type MGService interface {
	Create(ctx context.Context, req domain.ModelGroup) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]domain.ModelGroup, int64, error)
	Delete(ctx context.Context, id int64) (int64, error)
}

type groupService struct {
	repo repository.MGRepository
}

func NewMGService(repo repository.MGRepository) MGService {
	return &groupService{
		repo: repo,
	}
}

func (s *groupService) List(ctx context.Context, offset, limit int64) ([]domain.ModelGroup, int64, error) {
	var (
		total int64
		mgs   []domain.ModelGroup
		eg    errgroup.Group
	)
	eg.Go(func() error {
		var err error
		mgs, err = s.repo.List(ctx, offset, limit)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx)
		return err
	})
	if err := eg.Wait(); err != nil {
		return mgs, total, err
	}
	return mgs, total, nil
}

func (s *groupService) Create(ctx context.Context, req domain.ModelGroup) (int64, error) {
	return s.repo.CreateModelGroup(ctx, req)
}

func (s *groupService) Delete(ctx context.Context, id int64) (int64, error) {
	return s.repo.DeleteModelGroup(ctx, id)
}
