package ginx

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorCoder 定义了包含错误码和错误消息的错误接口
type ErrorCoder interface {
	error
	GetCode() int
	GetMsg() string
}

func handleError(ctx *gin.Context, err error, systemResult Result) {
	var errCoder ErrorCoder
	if errors.As(err, &errCoder) {
		slog.Warn("执行业务逻辑产生业务错误", slog.Any("err", err))
		ctx.PureJSON(http.StatusOK, Result{
			Code: errCoder.GetCode(),
			Msg:  errCoder.GetMsg(),
		})
		return
	}
	slog.Error("执行业务逻辑失败", slog.Any("err", err))
	ctx.PureJSON(http.StatusInternalServerError, systemResult)
}

func WrapBody[Req any](fn func(ctx *gin.Context, req Req) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			slog.Error("绑定参数失败", slog.Any("err", err))
			return
		}

		res, err := fn(ctx, req)
		if err != nil {
			handleError(ctx, err, res)
			return
		}

		ctx.JSON(http.StatusOK, res)
	}
}

func Wrap(fn func(ctx *gin.Context) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := fn(ctx)
		if err != nil {
			handleError(ctx, err, res)
			return
		}
		ctx.PureJSON(http.StatusOK, res)
	}
}

func WrapFinder(fn func(ctx *gin.Context) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := fn(ctx)
		if err != nil {
			handleError(ctx, err, res)
			return
		}
		ctx.PureJSON(http.StatusOK, res.Data)
	}
}

func WrapFinderBody[Req any](fn func(ctx *gin.Context, req Req) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			slog.Error("绑定参数失败", slog.Any("err", err))
			return
		}

		res, err := fn(ctx, req)
		if err != nil {
			handleError(ctx, err, res)
			return
		}
		ctx.PureJSON(http.StatusOK, res.Data)
	}
}

func Ws(fn func(ctx *gin.Context) error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := fn(ctx)
		if err != nil {
			slog.Warn("Websocket", slog.Any("err", err))
			return
		}
		ctx.PureJSON(http.StatusOK, "OK")
	}
}
