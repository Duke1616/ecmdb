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
	g.POST("/create", ginx.WrapBody[CreateResourceReq](h.CreateResource))

	// 查看资产详情信息
	g.POST("/detail", ginx.WrapBody[DetailResourceReq](h.DetailResource))

	// 根据模型查看资产列表
	g.POST("/list", ginx.WrapBody[ListResourceReq](h.ListResource))

	// 根据 ids 查询模型资产列表
	g.POST("/list/ids", ginx.WrapBody[ListResourceIdsReq](h.ListResourceByIds))

}

func (h *Handler) CreateResource(ctx *gin.Context, req CreateResourceReq) (ginx.Result, error) {
	id, err := h.svc.CreateResource(ctx, domain.Resource{
		Name:     req.Name,
		ModelUID: req.ModelUid,
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
	fields, err := h.attributeSvc.SearchAttributeFieldsByModelUid(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	resp, err := h.svc.FindResourceById(ctx, fields, req.ID)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: resp,
		Msg:  "查看资源详情成功",
	}, nil
}

func (h *Handler) ListResource(ctx *gin.Context, req ListResourceReq) (ginx.Result, error) {
	fields, err := h.attributeSvc.SearchAttributeFieldsByModelUid(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	resp, err := h.svc.ListResource(ctx, fields, req.ModelUid, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: resp,
		Msg:  "查看资源列表成功",
	}, nil
}

func (h *Handler) ListResourceByIds(ctx *gin.Context, req ListResourceIdsReq) (ginx.Result, error) {
	fields, err := h.attributeSvc.SearchAttributeFieldsByModelUid(ctx, req.ModelUid)
	if err != nil {
		return ginx.Result{}, err
	}

	rrs, err := h.svc.ListResourceByIds(ctx, fields, req.Ids)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{Data: rrs}, nil
}
