package web

import (
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/domain"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/service"
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

func (h *Handler) PublicRoutes(server *gin.Engine) {
	g := server.Group("/api/endpoint")
	g.POST("/register", ginx.WrapBody[RegisterEndpointReq](h.RegisterEndpoint))
}

func (h *Handler) RegisterEndpoint(ctx *gin.Context, req RegisterEndpointReq) (ginx.Result, error) {
	eId, err := h.svc.RegisterEndpoint(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: eId,
	}, nil
}

func (h *Handler) toDomain(req RegisterEndpointReq) domain.Endpoint {
	return domain.Endpoint{
		Path:         req.Path,
		Method:       req.Method,
		Resource:     req.Resource,
		Desc:         req.Desc,
		IsAuth:       req.IsAuth,
		IsAudit:      req.IsAudit,
		IsPermission: req.IsPermission,
	}
}
