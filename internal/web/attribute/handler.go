package web

import (
	"errors"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/errs"
	service "github.com/Duke1616/ecmdb/internal/service/attribute"
	modelservice "github.com/Duke1616/ecmdb/internal/service/model"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	svc      service.Service
	modelSvc modelservice.Service
	capability.IRegistry
}

func NewHandler(svc service.Service, modelSvc modelservice.Service) *Handler {
	return &Handler{
		svc:       svc,
		modelSvc:  modelSvc,
		IRegistry: capability.NewRegistry("cmdb", "attribute", "模型管理/属性管理"),
	}
}

// PrivateRoutes 注册属性管理模块需要中心化登录及权限判定（由 EIAM SDK 统一拦截承载）的私有路由
func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/attribute")

	// ==========================================
	// 1. 属性分组管理接口
	// ==========================================
	// 创建属性分组
	g.POST("/group/create", h.Capability("创建分组", "group_add").
		Handle(ginx.WrapBody[CreateAttributeGroup](h.CreateAttributeGroup)),
	)

	// 根据 ID 批量查询属性分组
	g.POST("/group/ids", h.Capability("批量查询分组", "group_view_by_ids").
		NoSync().
		Handle(ginx.WrapBody[ListAttributeGroupByIdsReq](h.ListAttributeGroupByIds)),
	)

	// 删除属性分组
	g.POST("/group/delete", h.Capability("删除分组", "group_delete").
		Handle(ginx.WrapBody[DeleteAttributeGroupReq](h.DeleteAttributeGroup)),
	)

	// 重命名属性分组
	g.POST("/group/rename", h.Capability("重命名分组", "group_rename").
		Handle(ginx.WrapBody[RenameAttributeGroupReq](h.RenameAttributeGroup)),
	)

	// 属性分组排序
	g.POST("/group/sort", h.Capability("分组排序", "group_sort").
		NoSync().
		Handle(ginx.WrapBody[SortAttributeGroupReq](h.SortAttributeGroup)),
	)

	// ==========================================
	// 2. 属性字段基础操作接口
	// ==========================================

	// 创建属性字段
	g.POST("/create", h.Capability("创建属性", "add").
		Handle(ginx.WrapBody[CreateAttributeReq](h.CreateAttribute)),
	)

	// 查询属性列表
	g.POST("/list", h.Capability("属性列表", "view").
		NoSync().
		Handle(ginx.WrapBody[ListAttributeReq](h.ListAttributes)),
	)

	// 查询属性字段列表
	g.POST("/list/field", h.Capability("属性字段", "view_fields").
		NoSync().
		Handle(ginx.WrapBody[ListAttributeReq](h.ListAttributeField)),
	)

	// 自定义属性列展示
	g.POST("/custom/field", h.Capability("自定义列展示", "view_custom_fields").
		Handle(ginx.WrapBody[CustomAttributeFieldColumnsReq](h.CustomAttributeFieldColumns)),
	)

	// 删除属性字段
	g.POST("/delete", h.Capability("删除属性", "delete").
		Handle(ginx.WrapBody[DeleteAttributeReq](h.DeleteAttribute)),
	)

	// 更新属性字段
	g.POST("/update", h.Capability("更新属性", "edit").
		Handle(ginx.WrapBody[UpdateAttributeReq](h.UpdateAttribute)),
	)

	// 属性字段排序
	g.POST("/sort", h.Capability("属性排序", "sort").
		Needs("cmdb:attribute:group_sort").
		Handle(ginx.WrapBody[SortAttributeReq](h.Sort)),
	)
}

func (h *Handler) CreateAttribute(ctx *gin.Context, req CreateAttributeReq) (ginx.Result, error) {
	id, err := h.svc.CreateAttribute(ctx.Request.Context(), toDomain(req))

	if mongo.IsDuplicateKeyError(err) {
		return duplicateErrorResult, fmt.Errorf("%w: %w", errs.ErrUniqueDuplicate, err)
	}

	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: id,
		Msg:  "添加模型属性成功",
	}, nil
}

func (h *Handler) UpdateAttribute(ctx *gin.Context, req UpdateAttributeReq) (ginx.Result, error) {
	id, err := h.svc.UpdateAttribute(ctx, h.toDomainUpdate(req))
	if err != nil {
		if errors.Is(err, errs.ErrConcurrentUpdate) {
			return ErrConcurrentUpdate, nil
		}
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
		Msg:  "更新模型属性成功",
	}, nil
}

func (h *Handler) ListAttributes(ctx *gin.Context, req ListAttributeReq) (ginx.Result, error) {
	model, err := h.modelSvc.GetByUid(ctx.Request.Context(), req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	groups, err := h.svc.ListAttributeGroup(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	attrs, _, err := h.svc.ListAttributes(ctx.Request.Context(), req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	fieldUIDsByGroup := make(map[int64][]string, len(groups))
	for _, attr := range attrs {
		fieldUIDsByGroup[attr.GroupId] = append(fieldUIDsByGroup[attr.GroupId], attr.FieldUid)
	}

	return ginx.Result{
		Data: RetrieveAttributeList{
			Model: AttributeModel{
				ModelUid: model.UID,
				Name:     model.Name,
			},
			Groups: slice.Map(groups, func(idx int, group domain.AttributeGroup) AttributeGroup {
				return AttributeGroup{
					GroupName: group.Name,
					ModelUid:  group.ModelUid,
					GroupId:   group.ID,
					Index:     group.SortKey,
					SortKey:   group.SortKey,
					FieldUids: fieldUIDsByGroup[group.ID],
				}
			}),
			Fields: slice.Map(attrs, func(idx int, attr domain.Attribute) Attribute {
				return toAttributeVo(attr)
			}),
		},
	}, nil
}

func (h *Handler) ListAttributeField(ctx *gin.Context, req ListAttributeReq) (ginx.Result, error) {
	attrs, total, err := h.svc.ListAttributes(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}
	att := slice.Map(attrs, func(idx int, src domain.Attribute) Attribute {
		return toAttributeVo(src)
	})

	return ginx.Result{
		Data: RetrieveAttributeFieldList{
			Total:      total,
			Attributes: att,
		},
	}, nil
}

func (h *Handler) CustomAttributeFieldColumns(ctx *gin.Context, req CustomAttributeFieldColumnsReq) (ginx.Result, error) {
	columns, err := h.svc.CustomAttributeFieldColumns(ctx.Request.Context(), req.ModelUid, req.CustomFieldName)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: columns,
	}, nil
}

func (h *Handler) DeleteAttribute(ctx *gin.Context, req DeleteAttributeReq) (ginx.Result, error) {
	count, err := h.svc.DeleteAttribute(ctx, req.Id)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
	}, nil
}

func (h *Handler) CreateAttributeGroup(ctx *gin.Context, req CreateAttributeGroup) (ginx.Result, error) {
	id, err := h.svc.CreateAttributeGroup(ctx.Request.Context(), h.toAttrGroupDomain(req))

	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: id,
		Msg:  "添加模型属性成功",
	}, nil
}

func (h *Handler) DeleteAttributeGroup(ctx *gin.Context, req DeleteAttributeGroupReq) (ginx.Result, error) {
	count, err := h.svc.DeleteAttributeGroup(ctx, req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
		Msg:  "删除属性分组成功",
	}, nil
}

func (h *Handler) RenameAttributeGroup(ctx *gin.Context, req RenameAttributeGroupReq) (ginx.Result, error) {
	_, err := h.svc.RenameAttributeGroup(ctx, req.ID, req.Name)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "重命名属性分组成功",
	}, nil
}

func (h *Handler) ListAttributeGroup(ctx *gin.Context, req ListAttributeGroupReq) (ginx.Result, error) {
	return ginx.Result{}, nil
}

func (h *Handler) ListAttributeGroupByIds(ctx *gin.Context, req ListAttributeGroupByIdsReq) (ginx.Result, error) {
	ags, err := h.svc.ListAttributeGroupByIds(ctx, req.Ids)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Code: 0,
		Msg:  "根据 ids 获取属性分组成功",
		Data: slice.Map(ags, func(idx int, src domain.AttributeGroup) AttributeGroup {
			return h.toAttrGroupVo(src)
		}),
	}, nil
}

func (h *Handler) toAttrGroupVo(src domain.AttributeGroup) AttributeGroup {
	return AttributeGroup{
		GroupName: src.Name,
		GroupId:   src.ID,
		Index:     src.SortKey,
		SortKey:   src.SortKey,
	}
}

func (h *Handler) toDomainUpdate(req UpdateAttributeReq) domain.Attribute {
	return domain.Attribute{
		ID:        req.Id,
		FieldName: req.FieldName,
		FieldType: req.FieldType,
		Required:  req.Required,
		Link:      req.Link,
		Secure:    req.Secure,
		Option:    req.Option,
		Index:     req.Index,
		SortKey:   req.SortKey,
	}
}

func (h *Handler) toAttrGroupDomain(req CreateAttributeGroup) domain.AttributeGroup {
	return domain.AttributeGroup{
		Name:     req.Name,
		ModelUid: req.ModelUid,
	}
}

// Sort 属性拖拽排序
func (h *Handler) Sort(ctx *gin.Context, req SortAttributeReq) (ginx.Result, error) {
	err := h.svc.Sort(ctx.Request.Context(), req.ID, req.TargetGroupID, req.TargetPosition)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "排序成功",
	}, nil
}

// SortAttributeGroup 属性组拖拽排序
func (h *Handler) SortAttributeGroup(ctx *gin.Context, req SortAttributeGroupReq) (ginx.Result, error) {
	err := h.svc.SortAttributeGroup(ctx.Request.Context(), req.ID, req.TargetPosition)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "排序成功",
	}, nil
}
