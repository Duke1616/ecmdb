package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Duke1616/ecmdb/internal/policy/internal/domain"
	"github.com/Duke1616/ecmdb/internal/policy/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc       service.Service
	sp        session.Provider
	threshold time.Duration
}

func NewHandler(svc service.Service, sp session.Provider) *Handler {
	return &Handler{
		svc:       svc,
		sp:        sp,
		threshold: time.Minute,
	}
}

// PublicRoutes 公开路由，供第三方 SDK 通过 HTTP 调用，不经过登录和鉴权中间件
func (h *Handler) PublicRoutes(server *gin.Engine) {
	g := server.Group("/api/policy")
	g.POST("/check_login", ginx.Wrap(h.CheckLoginForSDK))
	g.POST("/check_policy", ginx.WrapBody[CheckPolicyReq](h.CheckPolicyForSDK))
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/policy")
	g.POST("/add/p", ginx.WrapBody[PolicyReq](h.AddPolicies))
	g.POST("/update/p", ginx.WrapBody[PolicyReq](h.UpdatePolicies))
	g.POST("/add/g", ginx.WrapBody[AddGroupingPolicyReq](h.AddGroupingPolicy))
	g.POST("/authorize", ginx.WrapBody[AuthorizeReq](h.Authorize))
	g.POST("/user/permissions", ginx.WrapBody[GetPermissionsForUserReq](h.GetImplicitPermissionsForUser))
	g.POST("/role/permissions", ginx.WrapBody[GetPermissionsForRoleReq](h.GetPermissionsForRole))
}

// CheckLoginForSDK 供第三方 SDK 调用的登录验证接口
// NOTE: Token 通过 HTTP Authorization Header 透传，session.Provider 原生处理
func (h *Handler) CheckLoginForSDK(ctx *gin.Context) (ginx.Result, error) {
	gCtx := &gctx.Context{Context: ctx}
	sess, err := h.sp.Get(gCtx)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return ginx.Result{}, fmt.Errorf("登录验证失败: %w", err)
	}

	// 检测 Token 是否过期
	expiration := sess.Claims().Expiration
	now := time.Now().UnixMilli()
	if expiration <= now {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return ginx.Result{}, fmt.Errorf("token 已过期")
	}

	// Token 即将过期时自动续期，新 Token 会通过 Response Header 返回
	if expiration-now < h.threshold.Milliseconds() {
		_ = h.sp.RenewAccessToken(gCtx)
	}

	return ginx.Result{
		Data: CheckLoginResp{
			Uid: sess.Claims().Uid,
		},
	}, nil
}

// CheckPolicyForSDK 供第三方 SDK 调用的权限鉴权接口
// NOTE: Token 通过 HTTP Authorization Header 透传
func (h *Handler) CheckPolicyForSDK(ctx *gin.Context, req CheckPolicyReq) (ginx.Result, error) {
	gCtx := &gctx.Context{Context: ctx}
	sess, err := h.sp.Get(gCtx)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return ginx.Result{}, fmt.Errorf("登录验证失败: %w", err)
	}

	userId := strconv.FormatInt(sess.Claims().Uid, 10)
	result, err := h.svc.Authorize(ctx, userId, req.Path, req.Method, req.Resource)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: result,
	}, nil
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

	policies := slice.Map(resp, func(idx int, src domain.Policy) Policy {
		return Policy{
			Path:   src.Path,
			Method: src.Method,
			Effect: Effect(src.Effect),
		}
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
	result, err := h.svc.Authorize(ctx, req.UserId, req.Path, req.Method, req.Resource)
	if err != nil {
		return systemErrorResult, err
	}

	if !result.Allowed {
		return ginx.Result{
			Code: 0,
			Msg:  result.Reason,
			Data: result,
		}, nil
	}

	return ginx.Result{
		Code: 0,
		Msg:  result.Reason,
		Data: result,
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
