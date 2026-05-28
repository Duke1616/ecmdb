package web

import (
	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	svc service.Service
	capability.IRegistry
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{
		svc:       svc,
		IRegistry: capability.NewRegistry("cmdb", "attribute", "属性管理"),
	}
}

// PrivateRoutes 注册属性管理模块需要中心化登录及权限判定（由 EIAM SDK 统一拦截承载）的私有路由
func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/attribute")

	// ==========================================
	// 1. 属性分组管理接口
	// ==========================================

	// 创建属性分组
	g.POST("/group/create", h.Capability("创建属性分组", "group_add").
		Handle(ginx.WrapBody[CreateAttributeGroup](h.CreateAttributeGroup)),
	)

	// 查询属性分组列表
	g.POST("/group/list", h.Capability("查询属性分组列表", "group_list").
		Handle(ginx.WrapBody[ListAttributeGroupReq](h.ListAttributeGroup)),
	)

	// 根据 ID 批量查询属性分组
	g.POST("/group/ids", h.Capability("批量查询属性分组", "group_view_by_ids").
		Handle(ginx.WrapBody[ListAttributeGroupByIdsReq](h.ListAttributeGroupByIds)),
	)

	// 删除属性分组
	g.POST("/group/delete", h.Capability("删除属性分组", "group_delete").
		Handle(ginx.WrapBody[DeleteAttributeGroupReq](h.DeleteAttributeGroup)),
	)

	// 重命名属性分组
	g.POST("/group/rename", h.Capability("重命名属性分组", "group_rename").
		Handle(ginx.WrapBody[RenameAttributeGroupReq](h.RenameAttributeGroup)),
	)

	// ==========================================
	// 2. 属性字段基础操作接口
	// ==========================================

	// 创建属性字段
	g.POST("/create", h.Capability("创建属性字段", "add").
		Handle(ginx.WrapBody[CreateAttributeReq](h.CreateAttribute)),
	)

	// 查询属性列表
	g.POST("/list", h.Capability("查询属性列表", "view").
		Handle(ginx.WrapBody[ListAttributeReq](h.ListAttributes)),
	)

	// 查询属性字段列表
	g.POST("/list/field", h.Capability("查询属性字段列表", "view_fields").
		Handle(ginx.WrapBody[ListAttributeReq](h.ListAttributeField)),
	)

	// 自定义属性列展示
	g.POST("/custom/field", h.Capability("自定义属性列展示", "view_custom_fields").
		Handle(ginx.WrapBody[CustomAttributeFieldColumnsReq](h.CustomAttributeFieldColumns)),
	)

	// 删除属性字段
	g.POST("/delete", h.Capability("删除属性字段", "delete").
		Handle(ginx.WrapBody[DeleteAttributeReq](h.DeleteAttribute)),
	)

	// 更新属性字段
	g.POST("/update", h.Capability("更新属性字段", "edit").
		Handle(ginx.WrapBody[UpdateAttributeReq](h.UpdateAttribute)),
	)

	// ==========================================
	// 3. 属性排序接口
	// ==========================================

	// 属性字段排序
	g.POST("/sort", h.Capability("属性拖拽排序", "sort").
		Handle(ginx.WrapBody[SortAttributeReq](h.Sort)),
	)

	// 属性分组排序
	g.POST("/group/sort", h.Capability("属性分组拖拽排序", "group_sort").
		Handle(ginx.WrapBody[SortAttributeGroupReq](h.SortAttributeGroup)),
	)
}

func (h *Handler) CreateAttribute(ctx *gin.Context, req CreateAttributeReq) (ginx.Result, error) {
	id, err := h.svc.CreateAttribute(ctx.Request.Context(), toDomain(req))

	if mongo.IsDuplicateKeyError(err) {
		return duplicateErrorResult, err
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
	attr := h.toDomainUpdate(req)
	t, err := h.svc.UpdateAttribute(ctx, attr)

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: t,
	}, nil
}

func (h *Handler) ListAttributes(ctx *gin.Context, req ListAttributeReq) (ginx.Result, error) {
	groups, err := h.svc.ListAttributeGroup(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	pipelines, err := h.svc.ListAttributePipeline(ctx.Request.Context(), req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	pipelineMap := make(map[int64]domain.AttributePipeline, len(pipelines))
	for _, p := range pipelines {
		pipelineMap[p.GroupId] = p
	}

	attributeList := slice.Map(groups, func(idx int, group domain.AttributeGroup) AttributeList {
		item := AttributeList{
			GroupId:   group.ID,
			GroupName: group.Name,
			Index:     group.SortKey,
			Expanded:  true,
		}

		if p, ok := pipelineMap[group.ID]; ok {
			item.Total = p.Total
			item.Attributes = slice.Map(p.Attributes, func(_ int, attr domain.Attribute) Attribute {
				return toAttributeVo(attr)
			})
		}

		return item
	})

	return ginx.Result{
		Data: RetrieveAttributeList{
			AttributeList: attributeList,
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
