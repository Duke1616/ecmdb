package web

import (
	"log/slog"
	"net/http"

	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	finderGinx "github.com/Duke1616/vuefinder-go/pkg/ginx"
	"github.com/gin-gonic/gin"
)

func wrapFinder(fn func(ctx *gin.Context) (finderGinx.Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := fn(ctx)
		if err != nil {
			writeFinderError(ctx, res.Message, err)
			return
		}
		ctx.PureJSON(http.StatusOK, res.Data)
	}
}

func wrapFinderBuff(fn func(ctx *gin.Context) (finderGinx.Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := fn(ctx)
		if err != nil {
			writeFinderError(ctx, res.Message, err)
			return
		}
		ctx.String(http.StatusOK, "%s", res.Data)
	}
}

func wrapFinderBody[Req any](fn func(ctx *gin.Context, req Req) (finderGinx.Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			writeFinderBindError(ctx, err)
			return
		}

		res, err := fn(ctx, req)
		if err != nil {
			writeFinderError(ctx, res.Message, err)
			return
		}
		ctx.PureJSON(http.StatusOK, res.Data)
	}
}

func wrapFinderBuffBody[Req any](fn func(ctx *gin.Context, req Req) (finderGinx.Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			writeFinderBindError(ctx, err)
			return
		}

		res, err := fn(ctx, req)
		if err != nil {
			writeFinderError(ctx, res.Message, err)
			return
		}
		ctx.String(http.StatusOK, "%s", res.Data)
	}
}

func writeFinderError(ctx *gin.Context, msg string, err error) {
	slog.Error("执行 Finder 逻辑失败", slog.Any("err", err))
	if msg == "" {
		msg = err.Error()
	}
	ctx.PureJSON(http.StatusInternalServerError, ginx.Result{
		Code: errs.SystemError.Code,
		Msg:  msg,
		Data: nil,
	})
}

func writeFinderBindError(ctx *gin.Context, err error) {
	slog.Error("绑定 Finder 参数失败", slog.Any("err", err))
	ctx.PureJSON(http.StatusBadRequest, ginx.Result{
		Code: errs.SystemError.Code,
		Msg:  err.Error(),
		Data: nil,
	})
}
