package web

import (
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type RelationTypeHandler struct {
	svc service.RelationTypeService
}

func NewRelationTypeHandler(svc service.RelationTypeService) *RelationTypeHandler {
	return &RelationTypeHandler{
		svc: svc,
	}
}

func (h *RelationTypeHandler) RegisterRoute(server *gin.Engine) {
	g := server.Group("/relation/type")
	// 关联类型
	g.POST("/create", ginx.WrapBody[CreateRelationTypeReq](h.Create))
}

func (h *RelationTypeHandler) Create(ctx *gin.Context, req CreateRelationTypeReq) (ginx.Result, error) {
	return ginx.Result{}, nil
}
