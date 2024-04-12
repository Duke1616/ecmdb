package gctx

import (
	"github.com/ecodeclub/ekit"
	"github.com/gin-gonic/gin"
)

type Context struct {
	*gin.Context
}

func (c *Context) Param(key string) ekit.AnyValue {
	return ekit.AnyValue{
		Val: c.Context.Param(key),
	}
}

func (c *Context) Query(key string) ekit.AnyValue {
	return ekit.AnyValue{
		Val: c.Context.Query(key),
	}
}
