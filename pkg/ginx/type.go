package ginx

import "github.com/Duke1616/ecmdb/pkg/ginx/gctx"

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type Context = gctx.Context
