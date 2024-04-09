package web

import (
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/model")

	g.POST("/group/create", ginx.WrapBody[CreateModelGroupReq](h.CreateGroup))
}

func (h *Handler) CreateGroup(ctx *gin.Context, req CreateModelGroupReq) (ginx.Result, error) {

	return ginx.Result{}, nil
}

func (h *Handler) CreateModel(*gin.Context) {

}

func (h *Handler) CreateModelAttr() {

}
