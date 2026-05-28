package web

import (
	"sort"

	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type RelationResourceHandler struct {
	svc service.RelationResourceService
	capability.IRegistry
}

func NewRelationResourceHandler(svc service.RelationResourceService) *RelationResourceHandler {
	return &RelationResourceHandler{
		svc:       svc,
		IRegistry: capability.NewRegistry("cmdb", "resource", "资源管理"),
	}
}

// PrivateRoute 注册资源关系管理需要中心化登录及权限判定（由 EIAM SDK 统一拦截承载）的私有路由
func (h *RelationResourceHandler) PrivateRoute(server *gin.Engine) {
	g := server.Group("/api/resource/relation")

	// ==========================================
	// 1. 资源关联关系基础接口
	// ==========================================

	// 创建资源关联关系
	g.POST("/create", h.Capability("创建资产关系", "relation_create").
		Handle(ginx.WrapBody[CreateResourceRelationReq](h.CreateResourceRelation)),
	)

	// 查询源资产关联关系 (暂不使用，保留注册)
	g.POST("/list/src", h.Capability("查询源资产关系", "relation_list_src").
		Handle(ginx.WrapBody[ListResourceDiagramReq](h.ListSrcResource)),
	)

	// 查询目标资产关联关系 (暂不使用，保留注册)
	g.POST("/list/dst", h.Capability("查询目标资产关系", "relation_list_dst").
		Handle(ginx.WrapBody[ListResourceDiagramReq](h.ListDstResource)),
	)

	// ==========================================
	// 2. 资源关联关系聚合/Pipeline 接口
	// ==========================================

	// 源资产关系聚合查询
	g.POST("/pipeline/src", h.Capability("源资产关系聚合查询", "relation_pipeline_src").
		Handle(ginx.WrapBody[ListResourceDiagramReq](h.ListSrcAggregated)),
	)

	// 目标资产关系聚合查询
	g.POST("/pipeline/dst", h.Capability("目标资产关系聚合查询", "relation_pipeline_dst").
		Handle(ginx.WrapBody[ListResourceDiagramReq](h.ListDstAggregated)),
	)

	// 所有资产关系聚合查询
	g.POST("/pipeline/all", h.Capability("所有资产关系聚合查询", "relation_pipeline_all").
		Handle(ginx.WrapBody[ListResourceDiagramReq](h.ListAllAggregated)),
	)

	// ==========================================
	// 3. 资源关联关系删除接口
	// ==========================================

	// 删除资产关系
	g.POST("/delete", h.Capability("删除资产关系", "relation_delete").
		Handle(ginx.WrapBody[DeleteResourceRelationReq](h.DeleteResourceRelation)),
	)
}

func (h *RelationResourceHandler) CreateResourceRelation(ctx *gin.Context, req CreateResourceRelationReq) (ginx.Result, error) {
	resp, err := h.svc.CreateResourceRelation(ctx, domain.ResourceRelation{
		RelationName:     req.RelationName,
		SourceResourceID: req.SourceResourceID,
		TargetResourceID: req.TargetResourceID,
	})

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "创建资源关联关系成功",
		Data: resp,
	}, nil
}

func (h *RelationResourceHandler) ListSrcResource(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	rrs, total, err := h.svc.ListSrcResources(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveRelationResource{
			Total: total,
			ResourceRelations: slice.Map(rrs, func(idx int, src domain.ResourceRelation) ResourceRelation {
				return h.toResourceRelationVo(src)
			}),
		},
	}, nil
}

func (h *RelationResourceHandler) ListDstResource(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	rrs, total, err := h.svc.ListDstResources(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveRelationResource{
			Total: total,
			ResourceRelations: slice.Map(rrs, func(idx int, src domain.ResourceRelation) ResourceRelation {
				return h.toResourceRelationVo(src)
			}),
		},
	}, nil
}

func (h *RelationResourceHandler) ListSrcAggregated(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	agg, err := h.svc.ListSrcAggregated(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Data: slice.Map(agg, func(idx int, src domain.ResourceAggregatedAssets) RetrieveAggregatedAssets {
			return h.toAggregatedAssetsVo(src)
		}),
	}, nil
}

func (h *RelationResourceHandler) ListDstAggregated(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	agg, err := h.svc.ListDstAggregated(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Data: slice.Map(agg, func(idx int, src domain.ResourceAggregatedAssets) RetrieveAggregatedAssets {
			return h.toAggregatedAssetsVo(src)
		}),
	}, nil
}

func (h *RelationResourceHandler) ListAllAggregated(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	var (
		eg   errgroup.Group
		srcS []domain.ResourceAggregatedAssets
		dstS []domain.ResourceAggregatedAssets
	)

	eg.Go(func() error {
		var err error
		srcS, err = h.svc.ListSrcAggregated(ctx, req.ModelUid, req.ResourceId)
		return err
	})

	eg.Go(func() error {
		var err error
		dstS, err = h.svc.ListDstAggregated(ctx, req.ModelUid, req.ResourceId)
		return err
	})
	if err := eg.Wait(); err != nil {
		return systemErrorResult, err
	}
	result := append(srcS, dstS...)
	sort.Slice(result, func(i, j int) bool {
		// 根据需要的排序逻辑进行排序，这里假设你有一个字段可以用来排序，比如 id
		return result[i].Total < result[j].Total
	})

	return ginx.Result{
		Data: slice.Map(result, func(idx int, src domain.ResourceAggregatedAssets) RetrieveAggregatedAssets {
			return h.toAggregatedAssetsVo(src)
		}),
	}, nil
}

func (h *RelationResourceHandler) DeleteResourceRelation(ctx *gin.Context, req DeleteResourceRelationReq) (ginx.Result, error) {
	id, err := h.svc.DeleteResourceRelationByName(ctx, req.ResourceId, req.ModelUid, req.RelationName)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
	}, nil
}

func (h *RelationResourceHandler) toResourceRelationVo(src domain.ResourceRelation) ResourceRelation {
	return ResourceRelation{
		ID:               src.ID,
		SourceModelUID:   src.SourceModelUID,
		TargetModelUID:   src.TargetModelUID,
		SourceResourceID: src.SourceResourceID,
		TargetResourceID: src.TargetResourceID,
		RelationTypeUID:  src.RelationTypeUID,
		RelationName:     src.RelationName,
	}
}

func (h *RelationResourceHandler) toAggregatedAssetsVo(src domain.ResourceAggregatedAssets) RetrieveAggregatedAssets {
	return RetrieveAggregatedAssets{
		RelationName: src.RelationName,
		ModelUid:     src.ModelUid,
		Total:        src.Total,
		ResourceIds:  src.ResourceIds,
	}
}
