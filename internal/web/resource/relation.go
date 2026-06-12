package web

import (
	"sort"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

func (h *Handler) CreateResourceRelation(ctx *gin.Context, req CreateResourceRelationReq) (ginx.Result, error) {
	resp, err := h.RRSvc.CreateResourceRelation(ctx, domain.ResourceRelation{
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

func (h *Handler) ListSrcResource(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	rrs, total, err := h.RRSvc.ListSrcResources(ctx, req.ModelUid, req.ResourceId)
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

func (h *Handler) ListDstResource(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	rrs, total, err := h.RRSvc.ListDstResources(ctx, req.ModelUid, req.ResourceId)
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

func (h *Handler) ListSrcAggregated(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	agg, err := h.RRSvc.ListSrcAggregated(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Data: slice.Map(agg, func(idx int, src domain.ResourceAggregatedAssets) RetrieveAggregatedAssets {
			return h.toAggregatedAssetsVo(src)
		}),
	}, nil
}

func (h *Handler) ListDstAggregated(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	agg, err := h.RRSvc.ListDstAggregated(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Data: slice.Map(agg, func(idx int, src domain.ResourceAggregatedAssets) RetrieveAggregatedAssets {
			return h.toAggregatedAssetsVo(src)
		}),
	}, nil
}

func (h *Handler) ListAllAggregated(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	var (
		eg   errgroup.Group
		srcS []domain.ResourceAggregatedAssets
		dstS []domain.ResourceAggregatedAssets
	)

	eg.Go(func() error {
		var err error
		srcS, err = h.RRSvc.ListSrcAggregated(ctx, req.ModelUid, req.ResourceId)
		return err
	})

	eg.Go(func() error {
		var err error
		dstS, err = h.RRSvc.ListDstAggregated(ctx, req.ModelUid, req.ResourceId)
		return err
	})
	if err := eg.Wait(); err != nil {
		return systemErrorResult, err
	}
	result := append(srcS, dstS...)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Total < result[j].Total
	})

	return ginx.Result{
		Data: slice.Map(result, func(idx int, src domain.ResourceAggregatedAssets) RetrieveAggregatedAssets {
			return h.toAggregatedAssetsVo(src)
		}),
	}, nil
}

func (h *Handler) DeleteResourceRelation(ctx *gin.Context, req DeleteResourceRelationReq) (ginx.Result, error) {
	id, err := h.RRSvc.DeleteResourceRelationByName(ctx, req.ResourceId, req.ModelUid, req.RelationName)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
	}, nil
}


func (h *Handler) toAggregatedAssetsVo(src domain.ResourceAggregatedAssets) RetrieveAggregatedAssets {
	return RetrieveAggregatedAssets{
		RelationName: src.RelationName,
		ModelUid:     src.ModelUid,
		Total:        src.Total,
		ResourceIds:  src.ResourceIds,
	}
}
