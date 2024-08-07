package web

import (
	"github.com/Duke1616/ecmdb/internal/menu/internal/domain"
	"github.com/Duke1616/ecmdb/internal/menu/internal/service"
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
	g := server.Group("/api/menu")
	g.POST("/create", ginx.WrapBody[CreateMenuReq](h.CreateMenu))
}

func (h *Handler) CreateMenu(ctx *gin.Context, req CreateMenuReq) (ginx.Result, error) {
	eId, err := h.svc.CreateMenu(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: eId,
	}, nil
}

func (h *Handler) toDomain(req CreateMenuReq) domain.Menu {
	return domain.Menu{
		Pid:    req.Pid,
		Name:   req.Name,
		Path:   req.Path,
		Sort:   req.Sort,
		IsRoot: req.IsRoot,
		Type:   req.Type,
		Meta: domain.Meta{
			Title:       req.Meta.Title,
			IsHidden:    req.Meta.IsHidden,
			IsAffix:     req.Meta.IsAffix,
			IsKeepAlive: req.Meta.IsKeepAlive,
			Icon:        req.Meta.Icon,
		},
		EndpointIds: req.EndpointIds,
	}
}
