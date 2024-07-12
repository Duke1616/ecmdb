package web

import (
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
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
	g := server.Group("/api/order")
	g.POST("/create", ginx.WrapBody[CreateOrderReq](h.CreateOrder))
}

func (h *Handler) CreateOrder(ctx *gin.Context, req CreateOrderReq) (ginx.Result, error) {
	order, err := h.svc.CreateOrder(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: order,
	}, nil
}

func (h *Handler) toDomain(req CreateOrderReq) domain.Order {
	return domain.Order{
		CreateBy:   req.CreateBy,
		TemplateId: req.TemplateId,
		FlowId:     req.FlowId,
		Data:       req.Data,
		Status:     domain.START,
	}
}
