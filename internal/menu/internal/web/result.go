package web

import (
	"github.com/Duke1616/ecmdb/internal/menu/internal/errs"
	"github.com/Duke1616/ecmdb/pkg/ginx"
)

var (
	systemErrorResult = ginx.Result{
		Code: errs.SystemError.Code,
		Msg:  errs.SystemError.Msg,
	}
	menuHasChildrenResult = ginx.Result{
		Code: errs.MenuHasChildren.Code,
		Msg:  errs.MenuHasChildren.Msg,
	}
)
