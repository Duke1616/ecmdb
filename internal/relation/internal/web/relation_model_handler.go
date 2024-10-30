package web

import (
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

type RelationModelHandler struct {
	svc service.RelationModelService
}

func NewRelationModelHandler(svc service.RelationModelService) *RelationModelHandler {
	return &RelationModelHandler{
		svc: svc,
	}
}

func (h *RelationModelHandler) PrivateRoute(server *gin.Engine) {
	g := server.Group("/api/model")
	// 模型关联关系
	g.POST("/relation/create", ginx.WrapBody[CreateModelRelationReq](h.CreateModelRelation))

	// 查询模型拥有的所有关联信息
	g.POST("/relation/list", ginx.WrapBody[ListModelRelationReq](h.ListModelUIDRelation))

	g.POST("/relation/delete", ginx.WrapBody[DeleteModelRelationReq](h.DeleteModelRelation))
}

func (h *RelationModelHandler) CreateModelRelation(ctx *gin.Context, req CreateModelRelationReq) (ginx.Result, error) {
	id, err := h.svc.CreateModelRelation(ctx, toModelDomain(req))

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "创建模型关联关系成功",
		Data: id,
	}, nil
}

// ListModelUIDRelation 根据模型唯一索引名称，查询所有关联信息
func (h *RelationModelHandler) ListModelUIDRelation(ctx *gin.Context, req ListModelRelationReq) (ginx.Result, error) {
	relations, total, err := h.svc.ListModelUidRelation(ctx, req.Offset, req.Limit, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveRelationModels{
			Total: total,
			ModelRelations: slice.Map(relations, func(idx int, src domain.ModelRelation) ModelRelation {
				return h.toRelationVO(src)
			}),
		},
	}, nil
}

func (h *RelationModelHandler) DeleteModelRelation(ctx *gin.Context, req DeleteModelRelationReq) (ginx.Result, error) {
	id, err := h.svc.DeleteModelRelation(ctx, req.Id)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
	}, nil
}

func toModelDomain(req CreateModelRelationReq) domain.ModelRelation {
	return domain.ModelRelation{
		SourceModelUID:  req.SourceModelUID,
		TargetModelUID:  req.TargetModelUID,
		RelationTypeUID: req.RelationTypeUID,
		Mapping:         req.Mapping,
	}
}

func (h *RelationModelHandler) toRelationVO(m domain.ModelRelation) ModelRelation {
	return ModelRelation{
		ID:              m.ID,
		SourceModelUID:  m.SourceModelUID,
		TargetModelUID:  m.TargetModelUID,
		RelationTypeUID: m.RelationTypeUID,
		RelationName:    m.RelationName,
		Mapping:         m.Mapping,
	}
}
