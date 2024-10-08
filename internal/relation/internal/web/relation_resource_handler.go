package web

import (
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"sort"
	"strings"
)

type RelationResourceHandler struct {
	svc service.RelationResourceService
}

func NewRelationResourceHandler(svc service.RelationResourceService) *RelationResourceHandler {
	return &RelationResourceHandler{
		svc: svc,
	}
}

func (h *RelationResourceHandler) PrivateRoute(server *gin.Engine) {
	g := server.Group("/api/resource/relation")
	// 资源关联关系
	g.POST("/create", ginx.WrapBody[CreateResourceRelationReq](h.CreateResourceRelation))

	// TODO 暂不使用，没有根据 relationName 进行筛选，会返回所有的结果
	g.POST("/list/src", ginx.WrapBody[ListResourceDiagramReq](h.ListSrcResource))
	g.POST("/list/dst", ginx.WrapBody[ListResourceDiagramReq](h.ListDstResource))

	// 列表聚合处理、通过聚合处理，为前端友好展示
	g.POST("/pipeline/src", ginx.WrapBody[ListResourceDiagramReq](h.ListSrcAggregated))
	g.POST("/pipeline/dst", ginx.WrapBody[ListResourceDiagramReq](h.ListDstAggregated))
	g.POST("/pipeline/all", ginx.WrapBody[ListResourceDiagramReq](h.ListAllAggregated))

	g.POST("/delete", ginx.WrapBody[DeleteResourceRelationReq](h.DeleteResourceRelation))
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
			return RetrieveAggregatedAssets{
				RelationName: src.RelationName,
				ModelUid:     src.ModelUid,
				Total:        src.Total,
				ResourceIds:  src.ResourceIds,
			}
		}),
	}, nil
}

func (h *RelationResourceHandler) DeleteResourceRelation(ctx *gin.Context, req DeleteResourceRelationReq) (ginx.Result, error) {
	var (
		id  int64
		err error
	)

	rn := strings.Split(req.RelationName, "_")
	if rn[0] == req.ModelUid {
		id, err = h.svc.DeleteSrcRelation(ctx, req.ResourceId, req.ModelUid, req.RelationName)
	} else {
		id, err = h.svc.DeleteDstRelation(ctx, req.ResourceId, req.ModelUid, req.RelationName)
	}

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
