package web

import (
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
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

func (h *Handler) RegisterRoute(server *gin.Engine) {
	g := server.Group("/user")
	g.POST("/ldap/login", ginx.WrapBody[LoginLdapReq](h.LoginLdap))
}

func (h *Handler) LoginLdap(ctx *gin.Context, req LoginLdapReq) (ginx.Result, error) {
	profile, err := h.ldapSvc.Login(ctx, domain.User{
		Username: req.User,
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

	return ginx.Result{
		Data: user,
		Msg:  "登录用户成功",
	}, nil
}
