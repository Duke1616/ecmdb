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
	g.POST("/user/permissions", ginx.WrapBody[GetPermissionsForUserReq](h.GetImplicitPermissionsForUser))
	g.POST("/role/permissions", ginx.WrapBody[GetPermissionsForRoleReq](h.GetPermissionsForRole))
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

func (h *Handler) GetImplicitPermissionsForUser(ctx *gin.Context, req GetPermissionsForUserReq) (ginx.Result, error) {
	resp, err := h.svc.GetImplicitPermissionsForUser(ctx, req.UserId)
	if err != nil {
		return systemErrorResult, err
	}

	// 返回后端结构
	uniquePermissions := make(map[Policy]bool)
	var policies []Policy
	policies = slice.FilterMap(resp, func(idx int, src domain.Policy) (Policy, bool) {
		policy := Policy{
			Path:   src.Path,
			Method: src.Method,
			Effect: Effect(src.Effect),
		}

		if !uniquePermissions[policy] {
			uniquePermissions[policy] = true
			return policy, true
		}

		return policy, false
	})

	return ginx.Result{
		Msg: "获取用户权限成功",
		Data: RetrievePolicies{
			Policies: policies,
		},
	}, nil
}

func (h *Handler) GetPermissionsForRole(ctx *gin.Context, req GetPermissionsForRoleReq) (ginx.Result, error) {
	pers, err := h.svc.GetPermissionsForRole(ctx, req.RoleCode)
	if err != nil {
		return systemErrorResult, err
	}

	policies := slice.Map(pers, func(idx int, src domain.Policy) Policy {
		return Policy{
			Path:   src.Path,
			Method: src.Method,
			Effect: Effect(src.Effect),
		}
	})

	return ginx.Result{
		Msg: "获取角色权限成功",
		Data: RetrievePolicies{
			Policies: policies,
		},
	}, nil
}

func (h *Handler) UpdatePolicies(ctx *gin.Context, req PolicyReq) (ginx.Result, error) {
	ok, err := h.svc.CreateOrUpdateFilteredPolicies(ctx, h.toDomain(req))
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
		Msg:  "权限通过",
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
		RoleCode: req.RoleCode,
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
		RoleCode: req.RoleCode,
	}
}
