package web

import (
	"errors"

	"github.com/Duke1616/ecmdb/internal/domain"
	service "github.com/Duke1616/ecmdb/internal/service/relation"
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/ecodeclub/ginx"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

var _ ginx.Handler = &RelationTypeHandler{}

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

func (h *RelationTypeHandler) PublicRoutes(_ *gin.Engine) {}

func (h *RelationTypeHandler) IdentifyRoutes(_ *gin.Engine) {}

// PrivateRoutes 注册关系类型管理需要中心化登录及权限判定（由 EIAM SDK 统一拦截承载）的私有路由
func (h *RelationTypeHandler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/relation")

	// ==========================================
	// 1. 关系类型管理接口
	// ==========================================

	// 创建关联类型
	g.POST("/create", h.Define("创建关联类型", "add").
		Bind(ginx.B[CreateRelationTypeReq](h.Create)),
	)

	// 查询关联类型列表
	g.POST("/list", h.Define("关联类型列表", "view").
		Bind(ginx.B[Page](h.List)),
	)

	// 更新关联类型
	g.POST("/update", h.Define("更新关联类型", "edit").
		Bind(ginx.B[UpdateRelationTypeReq](h.Update)),
	)

	// 删除关联类型
	g.POST("/delete", h.Define("删除关联类型", "delete").
		Bind(ginx.B[DeleteRelationTypeReq](h.Delete)),
	)
}

func (h *RelationTypeHandler) Create(ctx *ginx.Context, req CreateRelationTypeReq) (ginx.Result, error) {
	id, err := h.svc.Create(ctx.Context, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "创建关联类型成功",
		Data: id,
	}, nil
}

func (h *RelationTypeHandler) List(ctx *ginx.Context, req Page) (ginx.Result, error) {
	rts, total, err := h.svc.List(ctx.Context, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询关联类型成功",
		Data: RetrieveRelationType{
			Total: total,
			RelationTypes: lo.Map(rts, func(src domain.RelationType, _ int) RelationType {
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

func (h *RelationTypeHandler) Update(ctx *ginx.Context, req UpdateRelationTypeReq) (ginx.Result, error) {
	_, err := h.svc.Update(ctx.Context, h.toUpdateDomain(req))
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{Msg: "更新关联类型成功"}, nil
}

func (h *RelationTypeHandler) Delete(ctx *ginx.Context, req DeleteRelationTypeReq) (ginx.Result, error) {
	_, err := h.svc.Delete(ctx.Context, req.Id)
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

