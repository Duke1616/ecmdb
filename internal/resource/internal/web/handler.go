package web

import (
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/resource")

	g.POST("/create", ginx.WrapBody[CreateResourceReq](h.CreateResource))
}

func (h *Handler) CreateResource(ctx *gin.Context, req CreateResourceReq) (ginx.Result, error) {
	fmt.Println(req)
	return ginx.Result{
		Data: req,
	}, nil
}
