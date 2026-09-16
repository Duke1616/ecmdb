package web

import (
	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/ecodeclub/ginx"
)

var (
	systemErrorResult = ginx.Result{
		Code: errs.SystemError.Code,
		Msg:  errs.SystemError.Msg,
	}

	duplicateResourceResult = ginx.Result{
		Code: errs.ErrUniqueDuplicate.Code,
		Msg:  "资产名称已存在，请勿重复创建",
	}
)

