package middleware

import (
	"net/http"
	"time"

	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"
)

type CheckLoginMiddlewareBuilder struct {
	threshold time.Duration
	logger    *elog.Component
	sp        session.Provider
}

func NewCheckLoginMiddlewareBuilder(sp session.Provider) *CheckLoginMiddlewareBuilder {
	return &CheckLoginMiddlewareBuilder{
		logger:    elog.DefaultLogger,
		threshold: time.Minute * 1,
		sp:        sp,
	}
}

func (b *CheckLoginMiddlewareBuilder) Build() gin.HandlerFunc {
	threshold := b.threshold.Milliseconds()
	return func(ctx *gin.Context) {
		gCtx := &gctx.Context{Context: ctx}
		sess, err := b.sp.Get(gCtx)
		if err != nil {
			b.logger.Error("未授权", elog.FieldErr(err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		expiration := sess.Claims().Expiration
		now := time.Now().UnixMilli()

		// 如果 token 已经过期，直接返回未授权
		if expiration <= now {
			b.logger.Error("token 已过期", elog.Int64("expiration", expiration), elog.Int64("now", now))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 只有当剩余时间小于阈值时才续期
		remainingTime := expiration - now
		if remainingTime < threshold {
			// 刷新一个token
			err = b.sp.RenewAccessToken(gCtx)
			if err != nil {
				b.logger.Warn("刷新 token 失败", elog.String("err", err.Error()))
			}
		}
		ctx.Set(session.CtxSessionKey, sess)
	}
}
