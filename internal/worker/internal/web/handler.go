package web

import (
	"github.com/Duke1616/ecmdb/internal/worker/internal/service"
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
	g := server.Group("/api/worker")
	g.POST("/list", ginx.WrapBody[ListWorkerReq](h.ListWorker))
}

func (h *Handler) ListWorker(ctx *gin.Context, req ListWorkerReq) (ginx.Result, error) {
	h.svc.ListWorker(ctx, req.Offset, req.Limit)
	return ginx.Result{}, nil
}
