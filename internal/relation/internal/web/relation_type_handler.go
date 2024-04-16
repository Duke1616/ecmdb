package web

import (
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
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
	g.POST("/list", ginx.WrapBody[Page](h.List))
}

func (h *RelationTypeHandler) Create(ctx *gin.Context, req CreateRelationTypeReq) (ginx.Result, error) {
	id, err := h.svc.Create(ctx, domain.RelationType{
		UID:            req.UID,
		Name:           req.Name,
		SourceDescribe: req.SourceDescribe,
		TargetDescribe: req.TargetDescribe,
	})
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "创建关联类型成功",
		Data: id,
	}, nil
}

func (h *RelationTypeHandler) List(ctx *gin.Context, req Page) (ginx.Result, error) {
	m, _, err := h.svc.List(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "查询关联类型成功",
		Data: m,
	}, nil
}
