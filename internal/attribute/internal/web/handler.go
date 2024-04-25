package web

import (
	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/model/attribute")

	g.POST("/create", ginx.WrapBody[CreateAttributeReq](h.CreateAttribute))
	g.POST("/detail", ginx.WrapBody[DetailAttributeReq](h.DetailAttribute))
	g.POST("/list", ginx.WrapBody[ListAttributeReq](h.ListAttribute))
}

func (h *Handler) CreateAttribute(ctx *gin.Context, req CreateAttributeReq) (ginx.Result, error) {
	id, err := h.svc.CreateAttribute(ctx.Request.Context(), domain.Attribute{
		Name:      req.Name,
		ModelUID:  req.ModelUID,
		FieldName: req.FieldName,
		FieldType: req.FieldType,
		Required:  req.Required,
	})

	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: id,
		Msg:  "添加模型属性成功",
	}, nil
}

func (h *Handler) DetailAttribute(ctx *gin.Context, req DetailAttributeReq) (ginx.Result, error) {
	attr, err := h.svc.SearchAttributeFieldsByModelUid(ctx, req.ModelUid)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Data: attr,
	}, nil
}

func (h *Handler) ListAttribute(ctx *gin.Context, req ListAttributeReq) (ginx.Result, error) {
	ats, err := h.svc.ListAttribute(ctx, req.ModelUID)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Data: ats,
	}, nil
}
