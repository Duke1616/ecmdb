package service

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository"
	"golang.org/x/sync/errgroup"
)

type MGService interface {
	// Create 创建模型分组
	Create(ctx context.Context, req domain.ModelGroup) (int64, error)

	// BatchCreate 批量创建模型分组
	BatchCreate(ctx context.Context, req []domain.ModelGroup) ([]domain.ModelGroup, error)

	// GetByNames 根据名称查询模型组
	GetByNames(ctx context.Context, names []string) ([]domain.ModelGroup, error)

	// GetByName 根据名称获取模型
	GetByName(ctx context.Context, name string) (domain.ModelGroup, error)

	// List 获取模型组列表
	List(ctx context.Context, offset, limit int64) ([]domain.ModelGroup, int64, error)

	// Delete 根据 ID 删除模型组
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

func (s *groupService) GetByName(ctx context.Context, name string) (domain.ModelGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (s *groupService) BatchCreate(ctx context.Context, req []domain.ModelGroup) ([]domain.ModelGroup, error) {
	return s.repo.BatchCreate(ctx, req)
}

func (s *groupService) GetByNames(ctx context.Context, names []string) ([]domain.ModelGroup, error) {
	return s.repo.GetByNames(ctx, names)
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
