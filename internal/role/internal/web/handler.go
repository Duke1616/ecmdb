package web

import (
	"github.com/Duke1616/ecmdb/internal/role/internal/domain"
	"github.com/Duke1616/ecmdb/internal/role/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
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
	g := server.Group("/api/role")
	g.POST("/update", ginx.WrapBody[UpdateRoleReq](h.UpdateRole))
	g.POST("/create", ginx.WrapBody[CreateRoleReq](h.CreateRole))
	g.POST("/list", ginx.WrapBody[Page](h.ListRole))
}

func (h *Handler) CreateRole(ctx *gin.Context, req CreateRoleReq) (ginx.Result, error) {
	rId, err := h.svc.CreateRole(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: rId,
	}, nil
}

func (h *Handler) UpdateRole(ctx *gin.Context, req UpdateRoleReq) (ginx.Result, error) {
	e := h.toDomainUpdate(req)

	t, err := h.svc.UpdateRole(ctx, e)

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: t,
	}, nil
}

func (h *Handler) ListRole(ctx *gin.Context, req Page) (ginx.Result, error) {
	rts, total, err := h.svc.ListRole(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询角色列表成功",
		Data: RetrieveRoles{
			Total: total,
			Roles: slice.Map(rts, func(idx int, src domain.Role) Role {
				return h.toVoRole(src)
			}),
		},
	}, nil
}

func (h *Handler) toDomain(req CreateRoleReq) domain.Role {
	return domain.Role{
		Name:   req.Name,
		Code:   req.Code,
		Desc:   req.Desc,
		Status: req.Status,
	}
}

func (h *Handler) toVoRole(req domain.Role) Role {
	return Role{
		Id:     req.Id,
		Name:   req.Name,
		Code:   req.Code,
		Desc:   req.Desc,
		Status: req.Status,
	}
}

func (h *Handler) toDomainUpdate(req UpdateRoleReq) domain.Role {
	return domain.Role{
		Id:     req.Id,
		Name:   req.Name,
		Code:   req.Code,
		Desc:   req.Desc,
		Status: req.Status,
	}
}
