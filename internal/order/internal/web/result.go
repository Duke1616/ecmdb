package web

import (
	"github.com/Duke1616/ecmdb/internal/order/internal/errs"
	"github.com/Duke1616/ecmdb/pkg/ginx"
)

var (
	systemErrorResult = ginx.Result{
		Code: errs.SystemError.Code,
		Msg:  errs.SystemError.Msg,
	}

	validateErrorResult = ginx.Result{
		Code: errs.ValidationError.Code,
		Msg:  errs.ValidationError.Msg,
	}
)
