package web

import (
	"encoding/json"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc     service.Service
	ldapSvc service.LdapService
}

func NewHandler(svc service.Service, ldapSvc service.LdapService) *Handler {
	return &Handler{
		svc:     svc,
		ldapSvc: ldapSvc,
	}
}

func (h *Handler) PublicRegisterRoutes(server *gin.Engine) {
	g := server.Group("/api/user")
	g.POST("/ldap/login", ginx.WrapBody[LoginLdapReq](h.LoginLdap))
	g.POST("/info", ginx.WrapBody[LoginLdapReq](h.Info))
	g.POST("/role/bind", ginx.WrapBody[UserBindRoleReq](h.UserRoleBind))
	g.POST("/refresh", ginx.Wrap(h.RefreshAccessToken))
	g.POST("/list", ginx.WrapBody[Page](h.ListUser))

}

func (h *Handler) LoginLdap(ctx *gin.Context, req LoginLdapReq) (ginx.Result, error) {
	profile, err := h.ldapSvc.Login(ctx, domain.User{
		Username: req.Username,
		Password: req.Password,
	})

	if err != nil {
		return systemErrorResult, err
	}

	user, err := h.svc.FindOrCreateByUsername(ctx, domain.User{
		Username:   profile.Username,
		Email:      profile.Email,
		Title:      profile.Title,
		SourceType: domain.Ldap,
		CreateType: domain.UserRegistry,
	})

	if err != nil {
		return systemErrorResult, err
	}

	u := h.ToUserVo(user)
	jwtData := make(map[string]string, 0)
	_, err = session.NewSessionBuilder(&gctx.Context{Context: ctx}, user.ID).SetJwtData(jwtData).Build()
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: u,
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
	bind, err := h.svc.AddRoleBind(ctx, req.Id, req.RoleCodes)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: bind,
	}, nil
}

func (h *Handler) RefreshAccessToken(ctx *gin.Context) (ginx.Result, error) {
	err := session.RenewAccessToken(&gctx.Context{Context: ctx})
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{Msg: "OK"}, nil
}

func (h *Handler) Info(ctx *gin.Context, req LoginLdapReq) (ginx.Result, error) {
	type AuthInfo struct {
		Username string   `json:"username"`
		Roles    []string `json:"roles"`
	}

	jsonData := `{"username":"admin","roles":["admin"]}`
	// 创建一个AuthInfo类型的变量来存储解析后的数据
	var authInfo AuthInfo

	// 使用json.Unmarshal函数解析JSON数据到结构体中
	json.Unmarshal([]byte(jsonData), &authInfo)
	return ginx.Result{
		Data: ginx.Result{Data: authInfo},
	}, nil
}

func (h *Handler) ToUserVo(src domain.User) User {
	return User{
		ID:         src.ID,
		Username:   src.Username,
		Email:      src.Email,
		Title:      src.Title,
		SourceType: src.SourceType,
		RoleCodes:  src.RoleCodes,
		CreateType: src.CreateType,
	}
}
