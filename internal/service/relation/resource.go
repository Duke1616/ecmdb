package service

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/internal/repository"
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

	// DeleteResourceRelationByName 根据关联名称和资源信息删除资产关联关系
	DeleteResourceRelationByName(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error)

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
	repo         repository.RelationResourceRepository
	modelRepo    repository.RelationModelRepository
	resourceRepo resourceNameRepository
}

func NewRelationResourceService(repo repository.RelationResourceRepository,
	modelRepo repository.RelationModelRepository,
	resourceRepo repository.ResourceRepository) RelationResourceService {
	return &resourceService{
		repo:         repo,
		modelRepo:    modelRepo,
		resourceRepo: resourceRepo,
	}
}

type resourceNameRepository interface {
	FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error)
}

func (s *resourceService) CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error) {
	// 1. 获取要校验的关系定义唯一标识名称
	relationName := req.RelationName
	if relationName == "" {
		relationName = fmt.Sprintf("%s_%s_%s", req.SourceModelUID, req.RelationTypeUID, req.TargetModelUID)
	}

	// 2. 从数据库加载拓扑规则定义
	mrs, err := s.modelRepo.GetByRelationNames(ctx, []string{relationName})
	if err != nil {
		return 0, fmt.Errorf("拓扑关联校验异常: %w", err)
	}
	if len(mrs) == 0 {
		return 0, relationValidationError(fmt.Errorf("拓扑关联校验失败：模型关系定义 %s 未注册", relationName))
	}

	// 3. 委派给领域模型进行自我充血式校验与数据补全
	if err = req.ValidateAndComplete(mrs[0]); err != nil {
		return 0, relationValidationError(err)
	}

	if err = s.checkMappingLimit(ctx, req, mrs[0]); err != nil {
		return 0, err
	}

	// 4. 流畅落库
	return s.repo.CreateResourceRelation(ctx, req)
}

func (s *resourceService) checkMappingLimit(ctx context.Context, req domain.ResourceRelation, mr domain.ModelRelation) error {
	switch mr.Mapping {
	case "", domain.MappingManyToMany:
		return nil
	case domain.MappingOneToOne:
		if err := s.checkSourceLimit(ctx, req); err != nil {
			return err
		}
		return s.checkTargetLimit(ctx, req)
	case domain.MappingOneToMany:
		return s.checkTargetLimit(ctx, req)
	default:
		return errs.RelationMappingConstraint.WithMsg(fmt.Sprintf("不支持的模型关联映射类型: %s", mr.Mapping))
	}
}

func (s *resourceService) checkSourceLimit(ctx context.Context, req domain.ResourceRelation) error {
	ids, err := s.repo.ListSrcRelated(ctx, req.SourceModelUID, req.RelationName, req.SourceResourceID)
	if err != nil {
		return fmt.Errorf("查询源端关联失败: %w", err)
	}
	if len(ids) > 0 {
		return errs.RelationMappingConstraint.WithMsg(
			fmt.Sprintf("关联映射约束冲突：源端资源%s已存在关联，不能重复绑定",
				s.resourceDisplay(ctx, req.SourceModelUID, req.SourceResourceID)),
		)
	}
	return nil
}

func (s *resourceService) checkTargetLimit(ctx context.Context, req domain.ResourceRelation) error {
	ids, err := s.repo.ListDstRelated(ctx, req.TargetModelUID, req.RelationName, req.TargetResourceID)
	if err != nil {
		return fmt.Errorf("查询目标端关联失败: %w", err)
	}
	if len(ids) > 0 {
		return errs.RelationMappingConstraint.WithMsg(
			fmt.Sprintf("关联映射约束冲突：目标端资源%s已存在关联，不能重复绑定",
				s.resourceDisplay(ctx, req.TargetModelUID, req.TargetResourceID)),
		)
	}
	return nil
}

func (s *resourceService) resourceDisplay(ctx context.Context, modelUID string, id int64) string {
	fallback := fmt.Sprintf("（模型：%s，ID：%d）", modelUID, id)
	if s.resourceRepo == nil {
		return fallback
	}

	resource, err := s.resourceRepo.FindResourceById(ctx, []string{"name"}, id)
	if err != nil || resource.Name == "" {
		return fallback
	}
	return fmt.Sprintf("「%s」（模型：%s，ID：%d）", resource.Name, modelUID, id)
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

// DeleteResourceRelationByName 根据关联名称和资源信息删除资产关联关系
func (s *resourceService) DeleteResourceRelationByName(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error) {
	// NOTE: 优先通过定义表获取正确的拓扑模型，无惧任何下划线分割，且完全规避了原本在 Web 层的 Split Bug
	mrs, err := s.modelRepo.GetByRelationNames(ctx, []string{relationName})
	if err != nil {
		return 0, fmt.Errorf("查询关联定义异常: %w", err)
	}
	if len(mrs) == 0 {
		return 0, fmt.Errorf("未找到对应的关联关系定义: %s", relationName)
	}

	// NOTE: 借助 Domain 领域对象的 IsSource 与 IsTarget 充血方法来判定删除方向，消除 Handler 层的业务污染
	mr := mrs[0]
	if mr.IsSource(modelUid) {
		return s.repo.DeleteSrcRelation(ctx, resourceId, modelUid, relationName)
	} else if mr.IsTarget(modelUid) {
		return s.repo.DeleteDstRelation(ctx, resourceId, modelUid, relationName)
	}

	return 0, fmt.Errorf("模型 UID %s 不属于关联关系 %s", modelUid, relationName)
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
