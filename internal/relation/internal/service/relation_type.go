package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
	"golang.org/x/sync/errgroup"
)

var ErrDependency = errors.New("exist dependency")

type RelationTypeService interface {
	// Create 创建关联类型
	Create(ctx context.Context, req domain.RelationType) (int64, error)

	// GetByUids 根据 UID 获取关联类型
	GetByUids(ctx context.Context, uids []string) ([]domain.RelationType, error)

	// BatchCreate 批量创建
	BatchCreate(ctx context.Context, rts []domain.RelationType) error

	// List 获取关联类型列表
	List(ctx context.Context, offset, limit int64) ([]domain.RelationType, int64, error)

	// Update 更新关联类型
	Update(ctx context.Context, req domain.RelationType) (int64, error)

	// Delete 删除关联类型
	Delete(ctx context.Context, id int64) (int64, error)
}

type service struct {
	repo         repository.RelationTypeRepository
	modelRepo    repository.RelationModelRepository
	resourceRepo repository.RelationResourceRepository
}

func NewRelationTypeService(repo repository.RelationTypeRepository,
	modelRepo repository.RelationModelRepository,
	resourceRepo repository.RelationResourceRepository) RelationTypeService {
	return &service{
		repo:         repo,
		modelRepo:    modelRepo,
		resourceRepo: resourceRepo,
	}
}

func (s *service) GetByUids(ctx context.Context, uids []string) ([]domain.RelationType, error) {
	return s.repo.GetByUids(ctx, uids)
}

func (s *service) BatchCreate(ctx context.Context, rts []domain.RelationType) error {
	return s.repo.BatchCreate(ctx, rts)
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

func (s *service) Update(ctx context.Context, req domain.RelationType) (int64, error) {
	return s.repo.Update(ctx, req)
}

func (s *service) Delete(ctx context.Context, id int64) (int64, error) {
	// 1. 查询关联类型
	rt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}

	// 2. 检查模型关联
	count, err := s.modelRepo.CountByRelationTypeUID(ctx, rt.UID)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, fmt.Errorf("%w: 关联类型正在被 %d 个模型关联使用，无法删除", ErrDependency, count)
	}

	// 3. 检查资源关联
	count, err = s.resourceRepo.CountByRelationTypeUID(ctx, rt.UID)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, fmt.Errorf("%w: 关联类型正在被 %d 个资源关联使用，无法删除", ErrDependency, count)
	}

	return s.repo.Delete(ctx, id)
}
