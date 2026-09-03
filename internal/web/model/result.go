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

	modelRelationIsNotFountResult = ginx.Result{
		Code: errs.RelationIsNotFountResult.Code,
		Msg:  errs.RelationIsNotFountResult.Msg,
	}
)

