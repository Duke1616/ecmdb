package web

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc          service.Service
	attributeSvc attribute.Service
}

func NewHandler(service service.Service, attributeSvc attribute.Service) *Handler {
	return &Handler{
		svc:          service,
		attributeSvc: attributeSvc,
	}
}

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/resource")
	g.POST("/create/:model_uid", ginx.WrapBody[CreateResourceReq](h.CreateResource))
	g.POST("/detail/:model_uid", ginx.WrapBody[DetailResourceReq](h.DetailResource))

	// 查询资源的关联关系
	g.POST("/list/relation", ginx.WrapBody[ListRelationsReq](h.ListRelations))
}

func (h *Handler) CreateResource(ctx *gin.Context, req CreateResourceReq) (ginx.Result, error) {
	modelUid := ctx.Param("model_uid")
	id, err := h.svc.CreateResource(ctx, domain.Resource{
		Name:     req.Name,
		ModelUID: modelUid,
		Data:     req.Data,
	})

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
		Msg:  "创建资源成功",
	}, nil
}

func (h *Handler) DetailResource(ctx *gin.Context, req DetailResourceReq) (ginx.Result, error) {
	modelUniqueIdentifier := ctx.Param("model_uid")
	attributes, err := h.attributeSvc.SearchAttributeByModelUID(ctx, modelUniqueIdentifier)
	if err != nil {
		return systemErrorResult, err
	}

	var dmAttr domain.DetailResource

	dmAttr.Projection = attributes
	dmAttr.ID = req.ID

	resp, err := h.svc.FindResourceById(ctx, dmAttr)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: resp,
		Msg:  "查看资源详情成功",
	}, nil
}

func (h *Handler) ListRelations(ctx *gin.Context, req ListRelationsReq) (ginx.Result, error) {
	return ginx.Result{}, nil
}
