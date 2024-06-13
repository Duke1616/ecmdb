package ginx

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

func WrapBody[Req any](fn func(ctx *gin.Context, req Req) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			slog.Error("绑定参数失败", slog.Any("err", err))
			return
		}

		res, err := fn(ctx, req)
		if err != nil {
			slog.Error("执行业务逻辑失败", slog.Any("err", err))
			ctx.PureJSON(http.StatusInternalServerError, res)
			return
		}

		ctx.JSON(http.StatusOK, res)
	}
}

func Wrap(fn func(ctx *gin.Context) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := fn(ctx)
		if err != nil {
			slog.Error("执行业务逻辑失败", slog.Any("err", err))
			ctx.PureJSON(http.StatusInternalServerError, res)
			return
		}
		ctx.PureJSON(http.StatusOK, res)
	}
}
