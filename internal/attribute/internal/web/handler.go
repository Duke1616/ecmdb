package web

import (
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

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/attribute")

	g.POST("/create", ginx.WrapBody[CreateAttributeReq](h.CreateAttribute))
	g.POST("/list", ginx.WrapBody[ListAttributeReq](h.ListAttributes))

}

func (h *Handler) CreateAttribute(ctx *gin.Context, req CreateAttributeReq) (ginx.Result, error) {
	id, err := h.svc.CreateAttribute(ctx.Request.Context(), toDomain(req))

	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: id,
		Msg:  "添加模型属性成功",
	}, nil
}

func (h *Handler) ListAttributes(ctx *gin.Context, req ListAttributeReq) (ginx.Result, error) {
	attrs, total, err := h.svc.ListAttributes(ctx, req.ModelUid)
	if err != nil {
		return systemErrorResult, err
	}

	att := slice.Map(attrs, func(idx int, src domain.Attribute) Attribute {
		return toAttributeVo(src)
	})
	atgroup1 := AttributeGroup{Attributes: att, GroupName: "字段1", GroupId: 1, Expanded: true}
	atgroup2 := AttributeGroup{Attributes: att, GroupName: "字段2", GroupId: 2, Expanded: true}
	var atgroups []AttributeGroup
	atgroups = append(atgroups, atgroup1)
	atgroups = append(atgroups, atgroup2)
	return ginx.Result{
		Data: RetrieveAttributeList{
			Total:      total,
			Attributes: atgroups,
		},
	}, nil
}
