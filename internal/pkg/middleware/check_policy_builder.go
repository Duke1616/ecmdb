package middleware

import (
	"net/http"
	"strconv"

	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"
)

type CheckPolicyMiddlewareBuilder struct {
	svc    policy.Service
	logger *elog.Component
	sp     session.Provider
}

func NewCheckPolicyMiddlewareBuilder(svc policy.Service, sp session.Provider) *CheckPolicyMiddlewareBuilder {
	return &CheckPolicyMiddlewareBuilder{
		svc:    svc,
		logger: elog.DefaultLogger,
		sp:     sp,
	}
}

func (c *CheckPolicyMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		gCtx := &ginx.Context{Context: ctx}
		sess, err := c.sp.Get(gCtx)
		if err != nil {
			gCtx.AbortWithStatus(http.StatusForbidden)
			c.logger.Debug("用户未登录", elog.FieldErr(err))
			return
		}

		// 获取请求的路径
		path := ctx.Request.URL.Path
		// 获取请求的HTTP方法
		method := ctx.Request.Method
		// 获取用户ID
		uid := sess.Claims().Uid
		ok, err := c.svc.Authorize(ctx.Request.Context(), strconv.FormatInt(uid, 10), path, method)
		if err != nil || !ok {
			gCtx.AbortWithStatus(http.StatusForbidden)
			c.logger.Debug("用户无权限", elog.FieldErr(err))
			return
		}
	}
}
