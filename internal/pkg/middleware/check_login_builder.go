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
		threshold: time.Minute * 60,
		sp:        sp,
	}
}

func (b *CheckLoginMiddlewareBuilder) Build() gin.HandlerFunc {
	threshold := b.threshold.Milliseconds()
	return func(ctx *gin.Context) {
		gCtx := &gctx.Context{Context: ctx}
		sess, err := b.sp.Get(gCtx)
		if err != nil {
			b.logger.Debug("未授权", elog.FieldErr(err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		expiration := sess.Claims().Expiration
		expiration = expiration - time.Now().UnixMilli()
		if expiration < threshold {
			err = b.sp.RenewAccessToken(gCtx)
			if err != nil {
				b.logger.Warn("刷新 token 失败", elog.FieldErr(err))
			}
		}
		ctx.Set(session.CtxSessionKey, sess)
	}
}
