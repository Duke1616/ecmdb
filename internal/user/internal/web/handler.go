package web

import (
	"fmt"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/service"
	"github.com/Duke1616/ecmdb/internal/user/ldapx"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc       service.Service
	ldapSvc   service.LdapService
	policySvc policy.Service
}

func NewHandler(svc service.Service, ldapSvc service.LdapService, policySvc policy.Service) *Handler {
	return &Handler{
		svc:       svc,
		ldapSvc:   ldapSvc,
		policySvc: policySvc,
	}
}

func (h *Handler) PublicRoutes(server *gin.Engine) {
	g := server.Group("/api/user")
	g.POST("/ldap/login", ginx.WrapBody[LoginLdapReq](h.LoginLdap))
	g.POST("/system/login", ginx.WrapBody[LoginSystemReq](h.LoginSystem))
	g.POST("/refresh", ginx.Wrap(h.RefreshAccessToken))
	g.POST("/register", ginx.WrapBody[RegisterUserReq](h.RegisterUser))
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/user")
	g.POST("/role/bind", ginx.WrapBody[UserBindRoleReq](h.UserRoleBind))
	g.POST("/list", ginx.WrapBody[Page](h.ListUser))
	g.POST("/info", ginx.Wrap(h.GetUserInfo))
}

func (h *Handler) LoginSystem(ctx *gin.Context, req LoginSystemReq) (ginx.Result, error) {
	user, err := h.svc.Login(ctx, req.Username, req.Password)
	if err != nil {
		return systemErrorResult, err
	}

	jwtData := make(map[string]string, 0)
	_, err = session.NewSessionBuilder(&gctx.Context{Context: ctx}, user.Id).SetJwtData(jwtData).Build()
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: h.ToUserVo(user),
		Msg:  "登录用户成功",
	}, nil
}

func (h *Handler) RegisterUser(ctx *gin.Context, req RegisterUserReq) (ginx.Result, error) {
	// 两次密码输入不一致
	if req.Password != req.RePassword {
		return systemErrorResult, nil
	}

	// 查询用户并创建
	u, err := h.svc.FindOrCreateBySystem(ctx, req.Username, req.Password, req.DisplayName)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: h.ToUserVo(u),
	}, nil
}

func (h *Handler) LoginLdap(ctx *gin.Context, req LoginLdapReq) (ginx.Result, error) {
	// LDAP 登录成功
	profile, err := h.ldapSvc.Login(ctx, domain.User{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return systemErrorResult, err
	}

	// 查找或插入用户
	user, err := h.svc.FindOrCreateByLdap(ctx, h.toDomain(profile))
	if err != nil {
		return systemErrorResult, err
	}

	// 生成session
	jwtData := make(map[string]string, 0)
	_, err = session.NewSessionBuilder(&gctx.Context{Context: ctx}, user.Id).SetJwtData(jwtData).Build()
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: h.ToUserVo(user),
		Msg:  "登录用户成功",
	}, nil
}

func (h *Handler) ListUser(ctx *gin.Context, req Page) (ginx.Result, error) {
	rts, total, err := h.svc.ListUser(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询用户列表成功",
		Data: RetrieveUsers{
			Total: total,
			Users: slice.Map(rts, func(idx int, src domain.User) User {
				return h.ToUserVo(src)
			}),
		},
	}, nil
}

func (h *Handler) UserRoleBind(ctx *gin.Context, req UserBindRoleReq) (ginx.Result, error) {
	_, err := h.svc.AddRoleBind(ctx, req.Id, req.RoleCodes)
	if err != nil {
		return systemErrorResult, err
	}

	ok, err := h.policySvc.UpdateFilteredGrouping(ctx, req.Id, req.RoleCodes)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Data: ok,
		Msg:  "用户角色绑定成功",
	}, nil
}

func (h *Handler) RefreshAccessToken(ctx *gin.Context) (ginx.Result, error) {
	err := session.RenewAccessToken(&gctx.Context{Context: ctx})
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{Msg: "OK"}, nil
}

func (h *Handler) GetUserInfo(ctx *gin.Context) (ginx.Result, error) {
	// 获取登录用户 sess 获取ID
	sess, err := session.Get(&gctx.Context{Context: ctx})
	if err != nil {
		return systemErrorResult, fmt.Errorf("获取 Session 失败, %w", err)
	}

	user, err := h.svc.FindById(ctx, sess.Claims().Uid)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Data: h.ToUserVo(user),
	}, nil
}

func (h *Handler) toDomain(profile *ldapx.Profile) domain.User {
	return domain.User{
		Username:    profile.Username,
		Email:       profile.Email,
		Title:       profile.Title,
		DisplayName: profile.DisplayName,
		Status:      domain.ENABLED,
		CreateType:  domain.LDAP,
	}
}

func (h *Handler) ToUserVo(src domain.User) User {
	return User{
		Id:          src.Id,
		Username:    src.Username,
		Email:       src.Email,
		Title:       src.Title,
		RoleCodes:   src.RoleCodes,
		DisplayName: src.DisplayName,
		CreateType:  src.CreateType.ToUint8(),
	}
}
