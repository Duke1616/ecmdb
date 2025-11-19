package service

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
	"golang.org/x/sync/errgroup"
)

//go:generate mockgen -source=./relation_model.go -destination=../../mocks/relation_model.mock.go -package=relationmocks -typed RelationModelService
type RelationModelService interface {
	// CreateModelRelation 创建模型关联关系
	CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error)

	// DeleteModelRelation 删除模型关联关系
	DeleteModelRelation(ctx context.Context, id int64) (int64, error)

	// ListModelUidRelation 根据模型 UID 获取。支持分页
	ListModelUidRelation(ctx context.Context, offset, limit int64, modelUid string) (
		[]domain.ModelRelation, int64, error)

	// CountByModelUid 根据模型 UID 获取数量
	CountByModelUid(ctx context.Context, modelUid string) (int64, error)

	// FindModelDiagramBySrcUids 查询模型关联关系，绘制拓扑图
	FindModelDiagramBySrcUids(ctx context.Context, srcUids []string) ([]domain.ModelDiagram, error)
}

type modelService struct {
	repo repository.RelationModelRepository
}

func (s *modelService) CountByModelUid(ctx context.Context, modelUid string) (int64, error) {
	return s.repo.TotalByModelUid(ctx, modelUid)
}

func NewRelationModelService(repo repository.RelationModelRepository) RelationModelService {
	return &modelService{
		repo: repo,
	}
}

func (s *modelService) CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error) {
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
	return s.repo.DeleteModelRelation(ctx, id)
}
