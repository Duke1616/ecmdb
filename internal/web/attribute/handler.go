package web

import (
	"errors"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/errs"
	service "github.com/Duke1616/ecmdb/internal/service/attribute"
	modelservice "github.com/Duke1616/ecmdb/internal/service/model"
	"github.com/Duke1616/ecmdb/pkg/contract/permission"
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/ecodeclub/ginx"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ ginx.Handler = &Handler{}

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

func (h *Handler) PublicRoutes(_ *gin.Engine) {}

func (h *Handler) IdentifyRoutes(_ *gin.Engine) {}

// PrivateRoutes 注册属性管理模块需要中心化登录及权限判定（由 EIAM SDK 统一拦截承载）的私有路由
func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/attribute")

	// ==========================================
	// 1. 属性分组管理接口 (派生属性分组层级)
	// ==========================================
	group := h.Sub("", "属性分组")

	// 创建属性分组
	g.POST("/group/create", group.Define("创建分组", "group_add").
		Bind(ginx.B[CreateAttributeGroup](h.CreateAttributeGroup)),
	)

	// 根据 ID 批量查询属性分组
	g.POST("/group/ids", group.Define("批量查询分组", "group_view_by_ids").
		NoSync().
		Bind(ginx.B[ListAttributeGroupByIdsReq](h.ListAttributeGroupByIds)),
	)

	// 删除属性分组
	g.POST("/group/delete", group.Define("删除分组", "group_delete").
		Bind(ginx.B[DeleteAttributeGroupReq](h.DeleteAttributeGroup)),
	)

	// 重命名属性分组
	g.POST("/group/rename", group.Define("重命名分组", "group_rename").
		Bind(ginx.B[RenameAttributeGroupReq](h.RenameAttributeGroup)),
	)

	// 属性分组排序
	g.POST("/group/sort", group.Define("分组排序", "group_sort").
		NoSync().
		Bind(ginx.B[SortAttributeGroupReq](h.SortAttributeGroup)),
	)


	// ==========================================
	// 2. 属性字段基础操作接口
	// ==========================================

	// 创建属性字段
	g.POST("/create", h.Define("创建属性", "add").
		Bind(ginx.B[CreateAttributeReq](h.CreateAttribute)),
	)

	// 查询属性列表
	g.POST("/list", h.Define("属性列表", "view").
		NoSync().
		Bind(ginx.B[ListAttributeReq](h.ListAttributes)),
	)

	// 查询属性字段列表
	g.POST("/list/field", h.Define("属性字段", "view_fields").
		NoSync().
		Bind(ginx.B[ListAttributeReq](h.ListAttributeField)),
	)

	// 自定义属性列展示
	g.POST("/custom/field", h.Define("自定义列展示", "view_custom_fields").
		Bind(ginx.B[CustomAttributeFieldColumnsReq](h.CustomAttributeFieldColumns)),
	)

	// 删除属性字段
	g.POST("/delete", h.Define("删除属性", "delete").
		Bind(ginx.B[DeleteAttributeReq](h.DeleteAttribute)),
	)

	// 更新属性字段
	g.POST("/update", h.Define("更新属性", "edit").
		Bind(ginx.B[UpdateAttributeReq](h.UpdateAttribute)),
	)

	// 属性字段排序
	g.POST("/sort", h.Define("属性排序", "sort").
		Needs(permission.Attribute.GroupSort).
		Bind(ginx.B[SortAttributeReq](h.Sort)),
	)
}




func (h *Handler) CreateAttribute(ctx *ginx.Context, req CreateAttributeReq) (ginx.Result, error) {
	id, err := h.svc.CreateAttribute(ctx.Context, toDomain(req))

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

func (h *Handler) UpdateAttribute(ctx *ginx.Context, req UpdateAttributeReq) (ginx.Result, error) {
	id, err := h.svc.UpdateAttribute(ctx.Context, h.toDomainUpdate(req))
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

func (h *Handler) ListAttributes(ctx *ginx.Context, req ListAttributeReq) (ginx.Result, error) {
	model, err := h.modelSvc.GetByUid(ctx.Context, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	groups, err := h.svc.ListAttributeGroup(ctx.Context, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	attrs, _, err := h.svc.ListAttributes(ctx.Context, req.ModelUid)
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
			Groups: lo.Map(groups, func(group domain.AttributeGroup, _ int) AttributeGroup {
				return AttributeGroup{
					GroupName: group.Name,
					ModelUid:  group.ModelUid,
					GroupId:   group.ID,
					Index:     group.SortKey,
					SortKey:   group.SortKey,
					FieldUids: fieldUIDsByGroup[group.ID],
				}
			}),
			Fields: lo.Map(attrs, func(attr domain.Attribute, _ int) Attribute {
				return toAttributeVo(attr)
			}),
		},
	}, nil
}

func (h *Handler) ListAttributeField(ctx *ginx.Context, req ListAttributeReq) (ginx.Result, error) {
	attrs, total, err := h.svc.ListAttributes(ctx.Context, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}
	att := lo.Map(attrs, func(src domain.Attribute, _ int) Attribute {
		return toAttributeVo(src)
	})

	return ginx.Result{
		Data: RetrieveAttributeFieldList{
			Total:      total,
			Attributes: att,
		},
	}, nil
}

func (h *Handler) CustomAttributeFieldColumns(ctx *ginx.Context, req CustomAttributeFieldColumnsReq) (ginx.Result, error) {
	columns, err := h.svc.CustomAttributeFieldColumns(ctx.Context, req.ModelUid, req.CustomFieldName)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: columns,
	}, nil
}

func (h *Handler) DeleteAttribute(ctx *ginx.Context, req DeleteAttributeReq) (ginx.Result, error) {
	count, err := h.svc.DeleteAttribute(ctx.Context, req.Id)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
	}, nil
}

func (h *Handler) CreateAttributeGroup(ctx *ginx.Context, req CreateAttributeGroup) (ginx.Result, error) {
	id, err := h.svc.CreateAttributeGroup(ctx.Context, h.toAttrGroupDomain(req))

	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: id,
		Msg:  "添加模型属性成功",
	}, nil
}

func (h *Handler) DeleteAttributeGroup(ctx *ginx.Context, req DeleteAttributeGroupReq) (ginx.Result, error) {
	count, err := h.svc.DeleteAttributeGroup(ctx.Context, req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
		Msg:  "删除属性分组成功",
	}, nil
}

func (h *Handler) RenameAttributeGroup(ctx *ginx.Context, req RenameAttributeGroupReq) (ginx.Result, error) {
	_, err := h.svc.RenameAttributeGroup(ctx.Context, req.ID, req.Name)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "重命名属性分组成功",
	}, nil
}

func (h *Handler) ListAttributeGroup(ctx *ginx.Context, req ListAttributeGroupReq) (ginx.Result, error) {
	return ginx.Result{}, nil
}

func (h *Handler) ListAttributeGroupByIds(ctx *ginx.Context, req ListAttributeGroupByIdsReq) (ginx.Result, error) {
	ags, err := h.svc.ListAttributeGroupByIds(ctx.Context, req.Ids)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Code: 0,
		Msg:  "根据 ids 获取属性分组成功",
		Data: lo.Map(ags, func(src domain.AttributeGroup, _ int) AttributeGroup {
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
func (h *Handler) Sort(ctx *ginx.Context, req SortAttributeReq) (ginx.Result, error) {
	err := h.svc.Sort(ctx.Context, req.ID, req.TargetGroupID, req.TargetPosition)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "排序成功",
	}, nil
}

// SortAttributeGroup 属性组拖拽排序
func (h *Handler) SortAttributeGroup(ctx *ginx.Context, req SortAttributeGroupReq) (ginx.Result, error) {
	err := h.svc.SortAttributeGroup(ctx.Context, req.ID, req.TargetPosition)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "排序成功",
	}, nil
}

