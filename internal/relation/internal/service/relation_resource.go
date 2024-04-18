package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
	"golang.org/x/sync/errgroup"
)

type RelationResourceService interface {
	CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error)
	ListResourceRelation(ctx context.Context, offset, limit int64) ([]domain.ResourceRelation, int64, error)

	// ListResourceIds 通过关联类型和模型UID 获取关联的 resources ids
	ListResourceIds(ctx context.Context, modelUid string, relationType string) ([]int64, error)

	// ListSrcResources 查询资源关联列表
	ListSrcResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, error)
	ListDstResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, error)

	// ListDiagram 通过 model_uid 和 resource_id 查询 SRC 和 DST 的数据
	ListDiagram(ctx context.Context, modelUid string, id int64) (domain.ResourceDiagram, int64, error)

	ListDstAggregated(ctx context.Context, modelUid string, id int64) (domain.ResourceAggregatedData, error)
	ListSrcAggregated(ctx context.Context, modelUid string, id int64) (domain.ResourceAggregatedData, error)
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

func (s *resourceService) ListResourceRelation(ctx context.Context, offset, limit int64) ([]domain.ResourceRelation, int64, error) {
	relation, err := s.repo.ListResourceRelation(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return relation, 0, nil
}

func (s *resourceService) ListResourceIds(ctx context.Context, modelUid string, relationType string) ([]int64, error) {
	return s.repo.ListResourceIds(ctx, modelUid, relationType)
}

func (s *resourceService) ListSrcResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, error) {
	return s.repo.ListSrcResources(ctx, modelUid, id)
}

func (s *resourceService) ListDstResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, error) {
	return s.repo.ListDstResources(ctx, modelUid, id)
}

func (s *resourceService) ListDiagram(ctx context.Context, modelUid string, id int64) (domain.ResourceDiagram, int64, error) {
	var (
		eg    errgroup.Group
		rd    domain.ResourceDiagram
		total int64
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

	total = int64(len(rd.SRC) + len(rd.DST))
	return rd, total, eg.Wait()
}

func (s *resourceService) ListSrcAggregated(ctx context.Context, modelUid string, id int64) (domain.ResourceAggregatedData, error) {
	return s.repo.ListSrcAggregated(ctx, modelUid, id)
}

func (s *resourceService) ListDstAggregated(ctx context.Context, modelUid string, id int64) (domain.ResourceAggregatedData, error) {
	return s.repo.ListDstAggregated(ctx, modelUid, id)
}
