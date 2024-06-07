package web

import (
	"github.com/Duke1616/ecmdb/internal/template/internal/service"
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
	g := server.Group("/api/template")
	// 模型分组
	g.POST("/create", ginx.WrapBody[CreateTemplateReq](h.CreateTemplate))
}

func (h *Handler) CreateTemplate(ctx *gin.Context, req CreateTemplateReq) (ginx.Result, error) {
	return ginx.Result{}, nil
}
