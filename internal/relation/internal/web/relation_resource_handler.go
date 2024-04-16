package web

import (
	"fmt"
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type RelationResourceHandler struct {
	svc          service.RelationResourceService
	attributeSvc attribute.Service
	resourceSvc  resource.Service
}

func NewRelationResourceHandler(svc service.RelationResourceService, attributeSvc attribute.Service,
	resourceSvc resource.Service) *RelationResourceHandler {
	return &RelationResourceHandler{
		svc:          svc,
		attributeSvc: attributeSvc,
		resourceSvc:  resourceSvc,
	}
}

func (h *RelationResourceHandler) RegisterRoute(server *gin.Engine) {
	g := server.Group("/relation/resource")
	// 资源关联关系
	g.POST("/create", ginx.WrapBody[CreateResourceRelationReq](h.CreateResourceRelation))
	g.POST("/list", ginx.WrapBody[Page](h.ListResourceRelation))
	g.POST("/list-name", ginx.WrapBody[ListResourceRelationByModelUidReq](h.ListResourceByModelUid))
}

func (h *RelationResourceHandler) CreateResourceRelation(ctx *gin.Context, req CreateResourceRelationReq) (ginx.Result, error) {
	resp, err := h.svc.CreateResourceRelation(ctx, domain.ResourceRelation{
		SourceModelUID:   req.SourceModelUID,
		TargetModelUID:   req.TargetModelUID,
		RelationTypeUID:  req.RelationTypeUID,
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

func (h *RelationResourceHandler) ListResourceByModelUid(ctx *gin.Context, req ListResourceRelationByModelUidReq) (
	ginx.Result, error) {
	projection, err := h.attributeSvc.SearchAttributeFiled(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, fmt.Errorf("查询字段属性失败: %w", err)
	}

	ids, err := h.svc.ListResourceIds(ctx, req.ModelUid, req.RelationType)
	if err != nil {
		return systemErrorResult, fmt.Errorf("查询 resource ids失败: %w", err)
	}

	resources, err := h.resourceSvc.ListResourceByIds(ctx, projection, ids)
	if err != nil {
		return systemErrorResult, fmt.Errorf("查询resources列表失败: %w", err)
	}

	return ginx.Result{
		Data: resources,
	}, nil
}
