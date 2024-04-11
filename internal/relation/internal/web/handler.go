package web

import (
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc         service.Service
	resourceSvc resource.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) RegisterRoute(server *gin.Engine) {
	g := server.Group("/relation")

	g.POST("/model/create", ginx.WrapBody[CreateModelRelationReq](h.CreateModelRelation))
	g.POST("/resource/create", ginx.WrapBody[CreateResourceRelationReq](h.CreateResourceRelation))
}

func (h *Handler) CreateModelRelation(ctx *gin.Context, req CreateModelRelationReq) (ginx.Result, error) {
	resp, err := h.svc.CreateModelRelation(ctx, domain.ModelRelation{
		SourceModelIdentifies:  req.SourceModelIdentifies,
		TargetModelIdentifies:  req.TargetModelIdentifies,
		RelationTypeIdentifies: req.RelationTypeIdentifies,
		Mapping:                req.Mapping,
	})

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "创建模型关联关系成功",
		Data: resp,
	}, nil
}

func (h *Handler) CreateResourceRelation(ctx *gin.Context, req CreateResourceRelationReq) (ginx.Result, error) {
	return ginx.Result{}, nil
}
