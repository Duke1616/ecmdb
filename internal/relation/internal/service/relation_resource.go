package service

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
	"golang.org/x/sync/errgroup"
)

//go:generate mockgen -source=./relation_resource.go -destination=../../mocks/relation_resource.mock.go -package=relationmocks -typed RelationResourceService
type RelationResourceService interface {
	CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error)

	// ListSrcResources 查询资源关联列表
	ListSrcResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, int64, error)
	ListDstResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, int64, error)

	// ListDiagram 通过 model_uid 和 resource_id 查询 SRC 和 DST 的数据
	ListDiagram(ctx context.Context, modelUid string, id int64) (domain.ResourceDiagram, int64, error)

	ListDstAggregated(ctx context.Context, modelUid string, id int64) ([]domain.ResourceAggregatedAssets, error)
	ListSrcAggregated(ctx context.Context, modelUid string, id int64) ([]domain.ResourceAggregatedAssets, error)

	// ListSrcRelated 查询当前已经关联的数据，新增资源关联使用
	ListSrcRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error)
	ListDstRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error)

	DeleteResourceRelation(ctx context.Context, id int64) (int64, error)

	DeleteSrcRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error)
	DeleteDstRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error)
}

type resourceService struct {
	repo repository.RelationResourceRepository
}

func NewRelationResourceService(repo repository.RelationResourceRepository) RelationResourceService {
	return &resourceService{
		repo: repo,
	}
}

func (s *resourceService) CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error) {
	return s.repo.CreateResourceRelation(ctx, req)
}

func (s *resourceService) ListSrcResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, int64, error) {
	var (
		eg    errgroup.Group
		total int64
		rrs   []domain.ResourceRelation
	)
	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalSrc(ctx, modelUid, id)
		return err
	})

	eg.Go(func() error {
		var err error
		rrs, err = s.repo.ListSrcResources(ctx, modelUid, id)
		return err
	})

	return rrs, total, eg.Wait()
}

func (s *resourceService) ListDstResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, int64, error) {
	var (
		eg    errgroup.Group
		total int64
		rrs   []domain.ResourceRelation
	)
	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalDst(ctx, modelUid, id)
		return err
	})

	eg.Go(func() error {
		var err error
		rrs, err = s.repo.ListDstResources(ctx, modelUid, id)
		return err
	})

	return rrs, total, eg.Wait()
}

func (s *resourceService) ListDiagram(ctx context.Context, modelUid string, id int64) (domain.ResourceDiagram, int64, error) {
	var (
		eg errgroup.Group
		rd domain.ResourceDiagram
	)
	eg.Go(func() error {
		var err error
		rd.SRC, err = s.repo.ListSrcResources(ctx, modelUid, id)
		return err
	})

	eg.Go(func() error {
		var err error
		rd.DST, err = s.repo.ListDstResources(ctx, modelUid, id)
		return err
	})

	return rd, int64(len(rd.SRC) + len(rd.DST)), eg.Wait()
}

func (s *resourceService) ListSrcAggregated(ctx context.Context, modelUid string, id int64) ([]domain.ResourceAggregatedAssets, error) {
	return s.repo.ListSrcAggregated(ctx, modelUid, id)
}

func (s *resourceService) ListDstAggregated(ctx context.Context, modelUid string, id int64) ([]domain.ResourceAggregatedAssets, error) {
	return s.repo.ListDstAggregated(ctx, modelUid, id)
}

func (s *resourceService) ListSrcRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error) {
	return s.repo.ListSrcRelated(ctx, modelUid, relationName, id)
}

func (s *resourceService) ListDstRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error) {
	return s.repo.ListDstRelated(ctx, modelUid, relationName, id)
}

func (s *resourceService) DeleteSrcRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error) {
	return s.repo.DeleteSrcRelation(ctx, resourceId, modelUid, relationName)
}

func (s *resourceService) DeleteDstRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error) {
	return s.repo.DeleteDstRelation(ctx, resourceId, modelUid, relationName)
}

func (s *resourceService) DeleteResourceRelation(ctx context.Context, id int64) (int64, error) {
	return s.repo.DeleteResourceRelation(ctx, id)
}
