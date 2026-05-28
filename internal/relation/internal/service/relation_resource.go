package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
	"golang.org/x/sync/errgroup"
)

// RelationResourceService 资源实例关联关系服务接口
//
//go:generate mockgen -source=./relation_resource.go -destination=../../mocks/relation_resource.mock.go -package=relationmocks -typed RelationResourceService
type RelationResourceService interface {
	// CreateResourceRelation 创建资源关联关系
	CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error)

	// ListSrcResources 查询资源关联列表
	ListSrcResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, int64, error)
	// ListDstResources 查询目标端关联资产列表
	ListDstResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, int64, error)

	// ListDiagram 通过 model_uid 和 resource_id 查询 SRC 和 DST 的数据
	ListDiagram(ctx context.Context, modelUid string, id int64) (domain.ResourceDiagram, int64, error)

	// ListDstAggregated 聚合查询目标关联列表
	ListDstAggregated(ctx context.Context, modelUid string, id int64) ([]domain.ResourceAggregatedAssets, error)
	// ListSrcAggregated 聚合查询源端关联列表
	ListSrcAggregated(ctx context.Context, modelUid string, id int64) ([]domain.ResourceAggregatedAssets, error)

	// ListSrcRelated 查询当前已经关联的数据，新增资源关联使用
	ListSrcRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error)

	// ListDstRelated 查询当前已经关联的目标数据，新增资源关联使用
	ListDstRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error)

	// DeleteResourceRelation 删除资源关联关系
	DeleteResourceRelation(ctx context.Context, id int64) (int64, error)

	// DeleteSrcRelation 删除源端关系
	DeleteSrcRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error)
	// DeleteDstRelation 删除目标端关系
	DeleteDstRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error)

	// CountByRelationTypeUID 根据关联类型 UID 获取数量
	CountByRelationTypeUID(ctx context.Context, uid string) (int64, error)

	// ListRecursiveDiagram 递归获取多级关联拓扑（支持最大深度）
	ListRecursiveDiagram(ctx context.Context, modelUid string, id int64, maxDepth int) (domain.ResourceDiagram, error)
}

type resourceService struct {
	repo      repository.RelationResourceRepository
	modelRepo repository.RelationModelRepository
}

func NewRelationResourceService(repo repository.RelationResourceRepository,
	modelRepo repository.RelationModelRepository) RelationResourceService {
	return &resourceService{
		repo:      repo,
		modelRepo: modelRepo,
	}
}

func (s *resourceService) CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error) {
	// 如果 RelationName 不为空但各个 Model/Type 字段为空（典型 REST API 入参），则进行 Split 解析补齐，确保数据落库完整性与拓扑校验通过
	// NOTE: 关系名称格式为: <源模型UID>_<关系类型UID>_<目标模型UID>
	if req.RelationName != "" && req.SourceModelUID == "" {
		parts := strings.Split(req.RelationName, "_")
		if len(parts) == 3 {
			req.SourceModelUID = parts[0]
			req.RelationTypeUID = parts[1]
			req.TargetModelUID = parts[2]
		}
	}

	// 拓扑一致性强校验：计算当前连线的 RelationName 并校验其在 ModelRelation 中是否存在
	relationName := fmt.Sprintf("%s_%s_%s", req.SourceModelUID, req.RelationTypeUID, req.TargetModelUID)

	// NOTE: 在保存前，我们还需要确保 req 的 RelationName 字段是正确的，以便后续查询
	if req.RelationName == "" {
		req.RelationName = relationName
	}

	mrs, err := s.modelRepo.GetByRelationNames(ctx, []string{req.RelationName})
	if err != nil {
		return 0, fmt.Errorf("拓扑关联校验异常: %w", err)
	}
	if len(mrs) == 0 {
		return 0, fmt.Errorf("拓扑关联校验失败：模型关系 %s -> %s (关系类型: %s) 未注册定义",
			req.SourceModelUID, req.TargetModelUID, req.RelationTypeUID)
	}

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

	if err := eg.Wait(); err != nil {
		return nil, 0, err
	}
	return rrs, total, nil
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

	if err := eg.Wait(); err != nil {
		return nil, 0, err
	}
	return rrs, total, nil
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

	if err := eg.Wait(); err != nil {
		return domain.ResourceDiagram{}, 0, err
	}
	return rd, int64(len(rd.SRC) + len(rd.DST)), nil
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

func (s *resourceService) CountByRelationTypeUID(ctx context.Context, uid string) (int64, error) {
	return s.repo.CountByRelationTypeUID(ctx, uid)
}

func (s *resourceService) ListRecursiveDiagram(ctx context.Context, modelUid string, id int64, maxDepth int) (domain.ResourceDiagram, error) {
	var (
		eg errgroup.Group
		rd domain.ResourceDiagram
	)

	eg.Go(func() error {
		var err error
		rd.SRC, err = s.repo.ListRecursiveSrc(ctx, modelUid, id, maxDepth)
		return err
	})

	eg.Go(func() error {
		var err error
		rd.DST, err = s.repo.ListRecursiveDst(ctx, modelUid, id, maxDepth)
		return err
	})

	if err := eg.Wait(); err != nil {
		return domain.ResourceDiagram{}, err
	}
	return rd, nil
}
