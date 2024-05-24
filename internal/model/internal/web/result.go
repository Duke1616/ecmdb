package web

import (
	"github.com/Duke1616/ecmdb/internal/model/internal/errs"
	"github.com/Duke1616/ecmdb/pkg/ginx"
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
