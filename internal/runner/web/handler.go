package web

import (
	"github.com/Duke1616/ecmdb/internal/runner/domain"
	"github.com/Duke1616/ecmdb/internal/runner/service"
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
	g := server.Group("/api/runner")
	g.POST("/publish", ginx.WrapBody[PublishRunnerReq](h.PublishRunner))
}

func (h *Handler) PublishRunner(ctx *gin.Context, req PublishRunnerReq) (ginx.Result, error) {
	h.svc.Publish(ctx, h.toDomain(req))

	return ginx.Result{
		Msg: "",
	}, nil
}

func (h *Handler) toDomain(req PublishRunnerReq) domain.Runner {
	return domain.Runner{
		Topic:    req.Topic,
		Name:     req.Name,
		UUID:     req.UUID,
		Language: req.Language,
		Code:     req.Code,
	}
}
