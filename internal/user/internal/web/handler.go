package web

import (
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	ldapSvc service.LdapService
}

func NewHandler(ldapSvc service.LdapService) *Handler {
	return &Handler{
		ldapSvc: ldapSvc,
	}
}

func (h *Handler) RegisterRoute(server *gin.Engine) {
	g := server.Group("/user")
	g.POST("/ldap/login", ginx.WrapBody[LoginLdapReq](h.LoginLdap))
}

func (h *Handler) LoginLdap(ctx *gin.Context, req LoginLdapReq) (ginx.Result, error) {
	username, err := h.ldapSvc.Login(ctx, domain.User{
		User:     req.User,
		Password: req.Password,
	})
	if err != nil {
		return ginx.Result{}, err
	}
	return ginx.Result{
		Data: username,
	}, nil
}
