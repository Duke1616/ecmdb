package service

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
	"golang.org/x/sync/errgroup"
)

//go:generate mockgen -source=./relation_model.go -destination=../../mocks/relation_model.mock.go -package=relationmocks -typed RelationModelService
type RelationModelService interface {
	// CreateModelRelation 创建模型关联关系
	CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error)

	// BatchCreate 批量创建模型关联关系
	BatchCreate(ctx context.Context, relations []domain.ModelRelation) error

	// DeleteModelRelation 删除模型关联关系
	DeleteModelRelation(ctx context.Context, id int64) (int64, error)

	// GetByRelationNames 根据唯一标识获取数据
	GetByRelationNames(ctx context.Context, names []string) ([]domain.ModelRelation, error)

	// ListModelUidRelation 根据模型 UID 获取。支持分页
	ListModelUidRelation(ctx context.Context, offset, limit int64, modelUid string) (
		[]domain.ModelRelation, int64, error)

	// CountByModelUid 根据模型 UID 获取数量
	CountByModelUid(ctx context.Context, modelUid string) (int64, error)

	// FindModelDiagramBySrcUids 查询模型关联关系，绘制拓扑图
	FindModelDiagramBySrcUids(ctx context.Context, srcUids []string) ([]domain.ModelDiagram, error)

	// CountByRelationTypeUID 根据关联类型 UID 获取数量
	CountByRelationTypeUID(ctx context.Context, uid string) (int64, error)
}

type modelService struct {
	repo         repository.RelationModelRepository
	resourceRepo repository.RelationResourceRepository
}

func (s *modelService) CountByModelUid(ctx context.Context, modelUid string) (int64, error) {
	return s.repo.TotalByModelUid(ctx, modelUid)
}

func NewRelationModelService(repo repository.RelationModelRepository, resourceRepo repository.RelationResourceRepository) RelationModelService {
	return &modelService{
		repo:         repo,
		resourceRepo: resourceRepo,
	}
}

func (s *modelService) BatchCreate(ctx context.Context, relations []domain.ModelRelation) error {
	return s.repo.BatchCreate(ctx, relations)
}

func (s *modelService) GetByRelationNames(ctx context.Context, names []string) ([]domain.ModelRelation, error) {
	return s.repo.GetByRelationNames(ctx, names)
}

func (s *modelService) CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error) {
	req.RelationName = req.RM()
	return s.repo.CreateModelRelation(ctx, req)
}

func (s *modelService) ListModelUidRelation(ctx context.Context, offset, limit int64, modelUid string) ([]domain.ModelRelation, int64, error) {
	var (
		eg    errgroup.Group
		mrs   []domain.ModelRelation
		total int64
	)
	eg.Go(func() error {
		var err error
		mrs, err = s.repo.ListRelationByModelUid(ctx, offset, limit, modelUid)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalByModelUid(ctx, modelUid)
		return err
	})

	if err := eg.Wait(); err != nil {
		return mrs, total, err
	}
	return mrs, total, nil
}

func (s *modelService) FindModelDiagramBySrcUids(ctx context.Context, srcUids []string) ([]domain.ModelDiagram, error) {
	return s.repo.FindModelDiagramBySrcUids(ctx, srcUids)
}

func (s *modelService) DeleteModelRelation(ctx context.Context, id int64) (int64, error) {
	mr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}

	count, err := s.resourceRepo.CountByRelationName(ctx, mr.RelationName)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, fmt.Errorf("%w: 关联关系正在被 %d 个资源关联数据使用，无法删除", ErrDependency, count)
	}

	return s.repo.DeleteModelRelation(ctx, id)
}

func (s *modelService) CountByRelationTypeUID(ctx context.Context, uid string) (int64, error) {
	return s.repo.CountByRelationTypeUID(ctx, uid)
}
