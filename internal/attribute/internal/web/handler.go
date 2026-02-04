package web

import (
	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	svc service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/attribute")
	// 字段分组
	g.POST("/group/create", ginx.WrapBody[CreateAttributeGroup](h.CreateAttributeGroup))
	g.POST("/group/list", ginx.WrapBody[ListAttributeGroupReq](h.ListAttributeGroup))
	g.POST("/group/ids", ginx.WrapBody[ListAttributeGroupByIdsReq](h.ListAttributeGroupByIds))
	g.POST("/group/delete", ginx.WrapBody[DeleteAttributeGroupReq](h.DeleteAttributeGroup))
	g.POST("/group/rename", ginx.WrapBody[RenameAttributeGroupReq](h.RenameAttributeGroup))

	// 字段操作
	g.POST("/create", ginx.WrapBody[CreateAttributeReq](h.CreateAttribute))
	g.POST("/list", ginx.WrapBody[ListAttributeReq](h.ListAttributes))
	g.POST("/list/field", ginx.WrapBody[ListAttributeReq](h.ListAttributeField))
	g.POST("/custom/field", ginx.WrapBody[CustomAttributeFieldColumnsReq](h.CustomAttributeFieldColumns))
	g.POST("/delete", ginx.WrapBody[DeleteAttributeReq](h.DeleteAttribute))
	g.POST("/update", ginx.WrapBody[UpdateAttributeReq](h.UpdateAttribute))

	// 属性排序
	g.POST("/sort", ginx.WrapBody[SortAttributeReq](h.Sort))
	g.POST("/group/sort", ginx.WrapBody[SortAttributeGroupReq](h.SortAttributeGroup))
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

	pipelines, err := h.svc.ListAttributePipeline(ctx, req.ModelUid)
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
	columns, err := h.svc.CustomAttributeFieldColumns(ctx, req.ModelUid, req.CustomFieldName)
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
		SortKey:  req.Index,
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
