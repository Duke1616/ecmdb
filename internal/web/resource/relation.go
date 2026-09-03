package web

import (
	"sort"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/ecodeclub/ginx"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

func (h *Handler) CreateResourceRelation(ctx *ginx.Context, req CreateResourceRelationReq) (ginx.Result, error) {
	resp, err := h.rrSvc.CreateResourceRelation(ctx.Context, domain.ResourceRelation{
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

func (h *Handler) ListAllAggregated(ctx *ginx.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	var (
		eg   errgroup.Group
		srcS []domain.ResourceAggregatedAssets
		dstS []domain.ResourceAggregatedAssets
	)

	eg.Go(func() error {
		var err error
		srcS, err = h.rrSvc.ListSrcAggregated(ctx.Context, req.ModelUid, req.ResourceId)
		return err
	})

	eg.Go(func() error {
		var err error
		dstS, err = h.rrSvc.ListDstAggregated(ctx.Context, req.ModelUid, req.ResourceId)
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
		Data: lo.Map(result, func(src domain.ResourceAggregatedAssets, _ int) RetrieveAggregatedAssets {
			return h.toAggregatedAssetsVo(src)
		}),
	}, nil
}

func (h *Handler) DeleteResourceRelation(ctx *ginx.Context, req DeleteResourceRelationReq) (ginx.Result, error) {
	id, err := h.rrSvc.DeleteResourceRelationByName(ctx.Context, req.ResourceId, req.ModelUid, req.RelationName)
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
