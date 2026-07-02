package web

import (
	"errors"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/service/relation"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

type RelationTypeHandler struct {
	svc service.RelationTypeService
	capability.IRegistry
}

func NewRelationTypeHandler(svc service.RelationTypeService) *RelationTypeHandler {
	return &RelationTypeHandler{
		svc:       svc,
		IRegistry: capability.NewRegistry("cmdb", "relation", "关联类型"),
	}
}

// PrivateRoute 注册关系类型管理需要中心化登录及权限判定（由 EIAM SDK 统一拦截承载）的私有路由
func (h *RelationTypeHandler) PrivateRoute(server *gin.Engine) {
	g := server.Group("/api/relation")

	// ==========================================
	// 1. 关系类型管理接口
	// ==========================================

	// 创建关联类型
	g.POST("/create", h.Capability("创建关联类型", "add").
		Handle(ginx.WrapBody[CreateRelationTypeReq](h.Create)),
	)

	// 查询关联类型列表
	g.POST("/list", h.Capability("关联类型列表", "view").
		Handle(ginx.WrapBody[Page](h.List)),
	)

	// 更新关联类型
	g.POST("/update", h.Capability("更新关联类型", "edit").
		Handle(ginx.WrapBody[UpdateRelationTypeReq](h.Update)),
	)

	// 删除关联类型
	g.POST("/delete", h.Capability("删除关联类型", "delete").
		Handle(ginx.WrapBody[DeleteRelationTypeReq](h.Delete)),
	)
}

func (h *RelationTypeHandler) Create(ctx *gin.Context, req CreateRelationTypeReq) (ginx.Result, error) {
	id, err := h.svc.Create(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "创建关联类型成功",
		Data: id,
	}, nil
}

func (h *RelationTypeHandler) List(ctx *gin.Context, req Page) (ginx.Result, error) {
	rts, total, err := h.svc.List(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询关联类型成功",
		Data: RetrieveRelationType{
			Total: total,
			RelationTypes: slice.Map(rts, func(idx int, src domain.RelationType) RelationType {
				return h.toRelationVO(src)
			}),
		},
	}, nil
}

func (h *RelationTypeHandler) toDomain(req CreateRelationTypeReq) domain.RelationType {
	return domain.RelationType{
		UID:            req.UID,
		Name:           req.Name,
		SourceDescribe: req.SourceDescribe,
		TargetDescribe: req.TargetDescribe,
	}
}

func (h *RelationTypeHandler) toRelationVO(src domain.RelationType) RelationType {
	return RelationType{
		ID:             src.ID,
		Name:           src.Name,
		UID:            src.UID,
		SourceDescribe: src.SourceDescribe,
		TargetDescribe: src.TargetDescribe,
	}
}

func (h *RelationTypeHandler) Update(ctx *gin.Context, req UpdateRelationTypeReq) (ginx.Result, error) {
	_, err := h.svc.Update(ctx, h.toUpdateDomain(req))
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{Msg: "更新关联类型成功"}, nil
}

func (h *RelationTypeHandler) Delete(ctx *gin.Context, req DeleteRelationTypeReq) (ginx.Result, error) {
	_, err := h.svc.Delete(ctx, req.Id)
	if err != nil {
		if errors.Is(err, service.ErrDependency) {
			return ginx.Result{
				Code: 501001,
				Msg:  err.Error(),
			}, nil
		}

		return systemErrorResult, err
	}
	return ginx.Result{Msg: "删除关联类型成功"}, nil
}

func (h *RelationTypeHandler) toUpdateDomain(req UpdateRelationTypeReq) domain.RelationType {
	return domain.RelationType{
		ID:             req.ID,
		Name:           req.Name,
		SourceDescribe: req.SourceDescribe,
		TargetDescribe: req.TargetDescribe,
	}
}
