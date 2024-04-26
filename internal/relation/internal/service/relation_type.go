package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
	"golang.org/x/sync/errgroup"
)

type RelationTypeService interface {
	Create(ctx context.Context, req domain.RelationType) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]domain.RelationType, int64, error)
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

func (s *service) List(ctx context.Context, offset, limit int64) ([]domain.RelationType, int64, error) {
	var (
		eg    errgroup.Group
		rts   []domain.RelationType
		total int64
	)
	eg.Go(func() error {
		var err error
		rts, err = s.repo.List(ctx, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx)
		return err
	})
	if err := eg.Wait(); err != nil {
		return rts, total, err
	}
	return rts, total, nil
}
