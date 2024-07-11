package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
	"github.com/Duke1616/ecmdb/internal/template/internal/repository"
	"golang.org/x/sync/errgroup"
)

type GroupService interface {
	Create(ctx context.Context, req domain.TemplateGroup) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]domain.TemplateGroup, int64, error)
}

type groupService struct {
	repo repository.TemplateGroupRepository
}

func NewGroupService(repo repository.TemplateGroupRepository) GroupService {
	return &groupService{
		repo: repo,
	}
}

func (s *groupService) Create(ctx context.Context, req domain.TemplateGroup) (int64, error) {
	return s.repo.Create(ctx, req)
}

func (s *groupService) List(ctx context.Context, offset, limit int64) ([]domain.TemplateGroup, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.TemplateGroup
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.List(ctx, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}
