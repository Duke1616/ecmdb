package web

import (
	"fmt"

	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
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

	// 字段操作
	g.POST("/create", ginx.WrapBody[CreateAttributeReq](h.CreateAttribute))
	g.POST("/list", ginx.WrapBody[ListAttributeReq](h.ListAttributes))
	g.POST("/list/field", ginx.WrapBody[ListAttributeReq](h.ListAttributeField))
	g.POST("/custom/field", ginx.WrapBody[CustomAttributeFieldColumnsReq](h.CustomAttributeFieldColumns))
	g.POST("/delete", ginx.WrapBody[DeleteAttributeReq](h.DeleteAttribute))
	g.POST("/update", ginx.WrapBody[UpdateAttributeReq](h.UpdateAttribute))
}

func (h *Handler) CreateAttribute(ctx *gin.Context, req CreateAttributeReq) (ginx.Result, error) {
	id, err := h.svc.CreateAttribute(ctx.Request.Context(), toDomain(req))

	fmt.Println(req)
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

	aps, err := h.svc.ListAttributePipeline(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}
	apsMap := make(map[int64]domain.AttributePipeline)
	for _, g := range aps {
		apsMap[g.GroupId] = g
	}

	attributeList := slice.Map(groups, func(idx int, src domain.AttributeGroup) AttributeList {
		mb := AttributeList{}
		val, ok := apsMap[src.ID]
		if ok {
			mb.GroupId = src.ID
			mb.GroupName = src.Name
			mb.Index = src.Index
			mb.Expanded = true
			mb.Total = val.Total
			mb.Attributes = slice.Map(val.Attributes, func(idx int, src domain.Attribute) Attribute {
				return toAttributeVo(src)
			})

			return mb
		}

		return AttributeList{
			GroupId:   src.ID,
			GroupName: src.Name,
			Expanded:  true,
			Index:     src.Index,
			Total:     0,
		}
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
		Index:     src.Index,
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
		Index:    req.Index,
	}
}
