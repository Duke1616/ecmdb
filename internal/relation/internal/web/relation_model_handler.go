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

func (h *RelationModelHandler) RegisterRoute(server *gin.Engine) {
	g := server.Group("/relation")
	// 模型关联关系
	g.POST("/model/create", ginx.WrapBody[CreateModelRelationReq](h.CreateModelRelation))
	g.POST("/model/list", ginx.WrapBody[Page](h.ListModelRelation))

	// 指定模型, 查询模型拥有的所有关联信息
	g.POST("/model/list-name", ginx.WrapBody[ListModelRelationByModelUidReq](h.ListModelUIDRelation))

	// TODO 模型匹配
	g.POST("/model/list-src", ginx.WrapBody[ListModelByUidReq](h.ListSrcModel))
	g.POST("/model/list-dst", ginx.WrapBody[ListModelByUidReq](h.ListDstModel))
}

func (h *RelationModelHandler) CreateModelRelation(ctx *gin.Context, req CreateModelRelationReq) (ginx.Result, error) {
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

func (h *RelationModelHandler) ListModelRelation(ctx *gin.Context, req Page) (ginx.Result, error) {
	m, _, err := h.svc.ListModelRelation(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "查询模型关联成功",
		Data: m,
	}, nil
}

// ListModelUIDRelation 根据模型唯一索引名称，查询所有关联信息
func (h *RelationModelHandler) ListModelUIDRelation(ctx *gin.Context, req ListModelRelationByModelUidReq) (ginx.Result, error) {
	relations, total, err := h.svc.ListModelUidRelation(ctx, req.Offset, req.Limit, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: ListRelationModelsResp{
			Total: total,
			ModelRelations: slice.Map(relations, func(idx int, src domain.ModelRelation) ModelRelation {
				return h.toRelationVO(src)
			}),
		},
	}, nil
}

func (h *RelationModelHandler) ListSrcModel(ctx *gin.Context, req ListModelByUidReq) (ginx.Result, error) {
	return ginx.Result{}, nil
}

func (h *RelationModelHandler) ListDstModel(ctx *gin.Context, req ListModelByUidReq) (ginx.Result, error) {
	return ginx.Result{}, nil
}

func (h *RelationModelHandler) toRelationVO(m domain.ModelRelation) ModelRelation {
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
