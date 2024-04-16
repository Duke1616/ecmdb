package web

import (
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

type RelationModelHandler struct {
	svc      service.RelationModelService
	modelSvc model.Service
}

func NewRelationModelHandler(svc service.RelationModelService, modelSvc model.Service) *RelationModelHandler {
	return &RelationModelHandler{
		svc:      svc,
		modelSvc: modelSvc,
	}
}

func (h *RelationModelHandler) RegisterRoute(server *gin.Engine) {
	g := server.Group("/relation")
	// 模型关联关系
	g.POST("/model/create", ginx.WrapBody[CreateModelRelationReq](h.CreateModelRelation))
	g.POST("/model/list", ginx.WrapBody[Page](h.ListModelRelation))
	g.POST("/model/list-name", ginx.WrapBody[ListModelRelationByModelUidReq](h.ListModelUIDRelation))

	// 查询所有模型的关联关系，拓补图
	g.POST("/model/diagram", ginx.WrapBody[Page](h.FindRelationModelDiagram))
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

func (h *RelationModelHandler) FindRelationModelDiagram(ctx *gin.Context, req Page) (ginx.Result, error) {
	//	1. 先查询所有的模型
	ms, _, err := h.modelSvc.ListModels(ctx, req.Offset, req.Limit)
	if err != nil {
		return ginx.Result{}, err
	}

	var diagrams []ModelDiagram

	for _, val := range ms {
		//  2. 以模型源作为基础，关联的所有服务
		rs, err := h.svc.FindModelRelationBySourceUID(ctx, val.UID)
		if err != nil {
			return systemErrorResult, err
		}

		var models []Model
		for _, rsVal := range rs {
			m := Model{
				ID:              rsVal.ID,
				RelationTypeUID: rsVal.RelationTypeUID,
				TargetModelUID:  rsVal.TargetModelUID,
			}

			models = append(models, m)
		}

		diagrams = append(diagrams, ModelDiagram{
			ID:        val.ID,
			ModelUID:  val.UID,
			ModelName: val.Name,
			Assets:    models,
		})
	}

	return ginx.Result{
		Data: RetrieveRelationModelDiagram{
			Diagrams: diagrams,
		},
	}, nil
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
