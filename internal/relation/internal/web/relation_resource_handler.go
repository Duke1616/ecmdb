package web

import (
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type RelationResourceHandler struct {
	svc service.RelationResourceService
}

func NewRelationResourceHandler(svc service.RelationResourceService) *RelationResourceHandler {
	return &RelationResourceHandler{
		svc: svc,
	}
}

func (h *RelationResourceHandler) RegisterRoute(server *gin.Engine) {
	g := server.Group("/resource/relation/")
	// 资源关联关系
	g.POST("/create", ginx.WrapBody[CreateResourceRelationReq](h.CreateResourceRelation))
	g.POST("/list/all", ginx.WrapBody[Page](h.ListResourceRelation))

	// 新建关联，查询所有的关联信息
	//g.POST("/list-name", ginx.WrapBody[ListResourceRelationByModelUidReq](h.ListResourceByModelUid))

	// 列表展示
	g.POST("/list-src", ginx.WrapBody[ListResourceDiagramReq](h.ListSrcResource))
	g.POST("/list-dst", ginx.WrapBody[ListResourceDiagramReq](h.ListDstResource))
	g.POST("/list", ginx.WrapBody[ListResourceDiagramReq](h.List))

	// 列表聚合展示、通过聚合处理
	g.POST("/pipeline/list-src", ginx.WrapBody[ListResourceDiagramReq](h.ListSrcAggregated))
	g.POST("/pipeline/list-dst", ginx.WrapBody[ListResourceDiagramReq](h.ListDstAggregated))
	g.POST("/pipeline/all", ginx.WrapBody[ListResourceDiagramReq](h.ListAllAggregated))
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

func (h *RelationResourceHandler) ListResourceRelation(ctx *gin.Context, req Page) (ginx.Result, error) {
	m, _, err := h.svc.ListResourceRelation(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "查询资源关联成功",
		Data: m,
	}, nil
}

// ListResourceByModelUid 查询模型下，可以进行关联的数据
//func (h *RelationResourceHandler) ListResourceByModelUid(ctx *gin.Context, req ListResourceRelationByModelUidReq) (
//	ginx.Result, error) {
//	fields, err := h.attributeSvc.SearchAttributeFieldsByModelUid(ctx, req.ModelUid)
//	if err != nil {
//		return systemErrorResult, fmt.Errorf("查询字段属性失败: %w", err)
//	}
//
//	ids, err := h.svc.ListResourceIds(ctx, req.ModelUid, req.RelationType)
//	if err != nil {
//		return systemErrorResult, fmt.Errorf("查询 resource ids失败: %w", err)
//	}
//
//	resources, err := h.resourceSvc.ListResourceByIds(ctx, fields, ids)
//	if err != nil {
//		return systemErrorResult, fmt.Errorf("查询resources列表失败: %w", err)
//	}
//
//	return ginx.Result{
//		Data: resources,
//	}, nil
//}

func (h *RelationResourceHandler) ListSrcResource(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	rs, err := h.svc.ListSrcResources(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: rs,
	}, nil
}

func (h *RelationResourceHandler) ListDstResource(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	rs, err := h.svc.ListDstResources(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: rs,
	}, nil
}

func (h *RelationResourceHandler) List(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	var (
		eg   errgroup.Group
		srcS []domain.ResourceRelation
		dstS []domain.ResourceRelation
	)

	eg.Go(func() error {
		var err error
		srcS, err = h.svc.ListSrcResources(ctx, req.ModelUid, req.ResourceId)
		return err
	})

	eg.Go(func() error {
		var err error
		dstS, err = h.svc.ListDstResources(ctx, req.ModelUid, req.ResourceId)
		return err
	})
	if err := eg.Wait(); err != nil {
		return systemErrorResult, err
	}
	result := append(srcS, dstS...)

	return ginx.Result{
		Data: result,
	}, nil
}

func (h *RelationResourceHandler) ListSrcAggregated(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	list, err := h.svc.ListSrcAggregated(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Data: list,
	}, nil
}

func (h *RelationResourceHandler) ListDstAggregated(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	list, err := h.svc.ListDstAggregated(ctx, req.ModelUid, req.ResourceId)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Data: list,
	}, nil
}

func (h *RelationResourceHandler) ListAllAggregated(ctx *gin.Context, req ListResourceDiagramReq) (ginx.Result, error) {
	var (
		eg   errgroup.Group
		srcS []domain.ResourceAggregatedData
		dstS []domain.ResourceAggregatedData
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
	return ginx.Result{
		Data: result,
	}, nil
}
