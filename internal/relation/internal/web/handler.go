package web

import (
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
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
	// 模型关联关系
	g.POST("/model/create", ginx.WrapBody[CreateModelRelationReq](h.CreateModelRelation))
	g.POST("/model/list", ginx.WrapBody[Page](h.ListModelRelation))
	g.POST("/model/list-name", ginx.WrapBody[ListModelRelationByModelUidReq](h.ListModelUIDRelation))

	// 资源关联关系
	g.POST("/resource/create", ginx.WrapBody[CreateResourceRelationReq](h.CreateResourceRelation))
	g.POST("/resource/list", ginx.WrapBody[Page](h.ListResourceRelation))
}

func (h *Handler) CreateModelRelation(ctx *gin.Context, req CreateModelRelationReq) (ginx.Result, error) {
	resp, err := h.svc.CreateModelRelation(ctx, domain.ModelRelation{
		SourceModelUID:  req.SourceModelUID,
		TargetModelUID:  req.TargetModelUID,
		RelationTypeUID: req.RelationTypeUID,
		Mapping:         req.Mapping,
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
	resp, err := h.svc.CreateResourceRelation(ctx, domain.ResourceRelation{
		SourceModelUID:   req.SourceModelUID,
		TargetModelUID:   req.TargetModelUID,
		RelationTypeUID:  req.RelationTypeIdentifies,
		SourceResourceID: req.SourceResourceID,
		TargetResourceID: req.TargetResourceID,
	})

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "创建资源关联关系成功",
		Data: resp,
	}, nil
}

func (h *Handler) ListModelRelation(ctx *gin.Context, req Page) (ginx.Result, error) {
	m, _, err := h.svc.ListModelRelation(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "查询模型关联成功",
		Data: m,
	}, nil
}

func (h *Handler) ListResourceRelation(ctx *gin.Context, req Page) (ginx.Result, error) {
	m, _, err := h.svc.ListResourceRelation(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "查询资源关联成功",
		Data: m,
	}, nil
}

// ListModelUIDRelation 根据模型唯一索引名称，查询所有关联信息
func (h *Handler) ListModelUIDRelation(ctx *gin.Context, req ListModelRelationByModelUidReq) (ginx.Result, error) {
	relations, total, err := h.svc.ListModelUidRelation(ctx, req.Offset, req.Limit, req.ModelIdentifies)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: ListOrdersResp{
			Total: total,
			ModelRelations: slice.Map(relations, func(idx int, src domain.ModelRelation) ModelRelation {
				return h.toRelationVO(src)
			}),
		},
	}, nil
}

func (h *Handler) toRelationVO(m domain.ModelRelation) ModelRelation {
	return ModelRelation{
		ID:              m.ID,
		SourceModelUID:  m.SourceModelUID,
		TargetModelUID:  m.TargetModelUID,
		RelationTypeUID: m.RelationTypeUID,
		RelationName:    m.RelationName,
		Mapping:         m.Mapping,
		Ctime:           m.Ctime,
		Utime:           m.Utime,
	}
}
