package web

import (
	"errors"

	"github.com/Duke1616/ecmdb/internal/domain"
	relationservice "github.com/Duke1616/ecmdb/internal/service/relation"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateModelRelation(ctx *gin.Context, req CreateModelRelationReq) (ginx.Result, error) {
	id, err := h.RMSvc.CreateModelRelation(ctx, toModelDomain(req))

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "创建模型关联关系成功",
		Data: id,
	}, nil
}

// ListModelUIDRelation 根据模型唯一索引名称，查询所有关联信息
func (h *Handler) ListModelUIDRelation(ctx *gin.Context, req ListModelRelationReq) (ginx.Result, error) {
	relations, total, err := h.RMSvc.ListModelUidRelation(ctx, req.Offset, req.Limit, req.ModelUid)
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

func (h *Handler) DeleteModelRelation(ctx *gin.Context, req DeleteModelRelationReq) (ginx.Result, error) {
	id, err := h.RMSvc.DeleteModelRelation(ctx, req.Id)
	if err != nil {
		if errors.Is(err, relationservice.ErrDependency) {
			return ginx.Result{
				Code: 501001,
				Msg:  err.Error(),
			}, nil
		}
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
	}, nil
}

func (h *Handler) UpdateModelRelation(ctx *gin.Context, req UpdateModelRelationReq) (ginx.Result, error) {
	_, err := h.RMSvc.UpdateModelRelation(ctx, toUpdateModelDomain(req))
	if err != nil {
		if errors.Is(err, relationservice.ErrDependency) {
			return ginx.Result{
				Code: 501001,
				Msg:  err.Error(),
			}, nil
		}
		return systemErrorResult, err
	}
	return ginx.Result{Msg: "更新模型关联关系成功"}, nil
}

func toUpdateModelDomain(req UpdateModelRelationReq) domain.ModelRelation {
	return domain.ModelRelation{
		ID:              req.ID,
		SourceModelUID:  req.SourceModelUID,
		TargetModelUID:  req.TargetModelUID,
		RelationTypeUID: req.RelationTypeUID,
		Mapping:         req.Mapping,
	}
}

func toModelDomain(req CreateModelRelationReq) domain.ModelRelation {
	return domain.ModelRelation{
		SourceModelUID:  req.SourceModelUID,
		TargetModelUID:  req.TargetModelUID,
		RelationTypeUID: req.RelationTypeUID,
		Mapping:         req.Mapping,
	}
}

func (h *Handler) toRelationVO(m domain.ModelRelation) ModelRelation {
	return ModelRelation{
		ID:              m.ID,
		SourceModelUID:  m.SourceModelUID,
		TargetModelUID:  m.TargetModelUID,
		RelationTypeUID: m.RelationTypeUID,
		RelationName:    m.RelationName,
		Mapping:         m.Mapping,
	}
}
