package web

import (
	"github.com/Duke1616/ecmdb/internal/task/errs"
	"github.com/Duke1616/ecmdb/pkg/ginx"
)

var (
	systemErrorResult = ginx.Result{
		Code: errs.SystemError.Code,
		Msg:  errs.SystemError.Msg,
	}

	validateErrorResult = ginx.Result{
		Code: errs.ValidateError.Code,
		Msg:  errs.ValidateError.Msg,
	}
)
