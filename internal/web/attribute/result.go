package web

import (
	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/pkg/ginx"
)

var (
	systemErrorResult = ginx.Result{
		Code: errs.SystemError.Code,
		Msg:  errs.SystemError.Msg,
	}

	duplicateErrorResult = ginx.Result{
		Code: 500001,
		Msg:  "唯一标识冲突",
	}

	ErrConcurrentUpdate = ginx.Result{
		Code: 40901,
		Msg:  "数据已被其他用户修改，请刷新后重试",
	}
)
