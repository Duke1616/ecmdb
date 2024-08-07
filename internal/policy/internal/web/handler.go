package web

import (
	"github.com/Duke1616/ecmdb/internal/policy/internal/domain"
	"github.com/Duke1616/ecmdb/internal/policy/internal/service"
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
	g := server.Group("/api/policy")
	g.POST("/add/p", ginx.WrapBody[PolicyReq](h.AddPolicies))
	g.POST("/update/p", ginx.WrapBody[PolicyReq](h.UpdatePolicies))
	g.POST("/add/g", ginx.WrapBody[AddGroupingPolicyReq](h.AddGroupingPolicy))
	g.POST("/authorize", ginx.WrapBody[AuthorizeReq](h.Authorize))
}

func (h *Handler) AddPolicies(ctx *gin.Context, req PolicyReq) (ginx.Result, error) {
	ok, err := h.svc.AddPolicies(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: ok,
	}, nil
}

func (h *Handler) UpdatePolicies(ctx *gin.Context, req PolicyReq) (ginx.Result, error) {
	ok, err := h.svc.UpdateFilteredPolicies(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: ok,
	}, nil
}

func (h *Handler) Authorize(ctx *gin.Context, req AuthorizeReq) (ginx.Result, error) {
	authorize, err := h.svc.Authorize(ctx, req.UserId, req.Path, req.Method)
	if err != nil {
		return systemErrorResult, err
	}

	if !authorize {
		return ginx.Result{
			Code: 0,
			Msg:  "权限拒绝",
			Data: authorize,
		}, nil
	}

	return ginx.Result{
		Code: 0,
		Msg:  "",
		Data: authorize,
	}, nil
}

func (h *Handler) AddGroupingPolicy(ctx *gin.Context, req AddGroupingPolicyReq) (ginx.Result, error) {
	ok, err := h.svc.AddGroupingPolicy(ctx, h.toDomainGroup(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: ok,
	}, nil
}

func (h *Handler) toDomain(req PolicyReq) domain.Policies {
	return domain.Policies{
		RoleName: req.RoleName,
		Policies: slice.Map(req.Policies, func(idx int, src Policy) domain.Policy {
			return domain.Policy{
				Path:   src.Path,
				Method: src.Method,
				Effect: domain.Effect(src.Effect),
			}
		}),
	}
}

func (h *Handler) toDomainGroup(req AddGroupingPolicyReq) domain.AddGroupingPolicy {
	return domain.AddGroupingPolicy{
		UserId:   req.UserId,
		RoleName: req.RoleName,
	}
}
