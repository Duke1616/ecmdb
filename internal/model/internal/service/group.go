package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository"
	"golang.org/x/sync/errgroup"
)

var ErrDependency = errors.New("存在关联数据")

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

	// Rename 根据 ID 重命名模型组
	Rename(ctx context.Context, id int64, name string) (int64, error)
}

type groupService struct {
	repo      repository.MGRepository
	modelRepo repository.ModelRepository
}

func NewMGService(repo repository.MGRepository, modelRepo repository.ModelRepository) MGService {
	return &groupService{
		repo:      repo,
		modelRepo: modelRepo,
	}
}

func (s *groupService) GetByName(ctx context.Context, name string) (domain.ModelGroup, error) {
	return s.repo.GetByName(ctx, name)
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
	count, err := s.modelRepo.CountByGroupId(ctx, id)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, fmt.Errorf("%w: 模型组正在被 %d 个模型使用，无法删除", ErrDependency, count)
	}
	return s.repo.DeleteModelGroup(ctx, id)
}

func (s *groupService) Rename(ctx context.Context, id int64, name string) (int64, error) {
	return s.repo.Rename(ctx, id, name)
}
