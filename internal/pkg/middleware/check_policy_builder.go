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
		sess, err := c.sp.Get(&ginx.Context{Context: ctx})
		if err != nil {
			c.logger.Warn("用户未登录", elog.FieldErr(err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 获取请求元数据
		path := ctx.Request.URL.Path
		method := ctx.Request.Method
		uid := sess.Claims().Uid

		// 调用鉴权服务，目前系统资源统一标识为 "CMDB"
		result, err := c.svc.Authorize(ctx.Request.Context(), strconv.FormatInt(uid, 10), path, method, "CMDB")
		if err != nil {
			c.logger.Error("权限鉴权服务异常", elog.FieldErr(err), elog.Int64("uid", uid), elog.String("path", path))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if !result.Allowed {
			c.logger.Warn("用户访问被拒绝",
				elog.Int64("uid", uid),
				elog.String("path", path),
				elog.String("method", method),
				elog.String("reason", result.Reason),
				elog.Any("roles", result.Roles),
				elog.Any("matched_policies", result.MatchedPolicies))
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		ctx.Next()
	}
}
