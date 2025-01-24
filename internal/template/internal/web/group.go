package web

import (
	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
	"github.com/Duke1616/ecmdb/internal/template/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	svc service.GroupService
}

func NewGroupHandler(svc service.GroupService) *GroupHandler {
	return &GroupHandler{
		svc: svc,
	}
}

func (h *GroupHandler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/template/group")
	g.POST("/create", ginx.WrapBody[CreateTemplateGroupReq](h.CreateTemplateGroup))
	g.POST("/list", ginx.WrapBody[Page](h.ListTemplateGroup))
	g.POST("/by_ids", ginx.WrapBody[FindTemplateGroupsByIdsReq](h.FindByIds))
}

func (h *GroupHandler) CreateTemplateGroup(ctx *gin.Context, req CreateTemplateGroupReq) (ginx.Result, error) {
	t, err := h.svc.Create(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: t,
	}, nil
}

func (h *GroupHandler) FindByIds(ctx *gin.Context, req FindTemplateGroupsByIdsReq) (ginx.Result, error) {
	rts, err := h.svc.ListByIds(ctx, req.Ids)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "根据 IDS 查询工单模版组成功",
		Data: RetrieveTemplateGroup{
			TemplateGroups: slice.Map(rts, func(idx int, src domain.TemplateGroup) TemplateGroup {
				return h.toVo(src)
			}),
		},
	}, nil
}

func (h *GroupHandler) ListTemplateGroup(ctx *gin.Context, req Page) (ginx.Result, error) {
	rts, total, err := h.svc.List(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询工单模版组列表成功",
		Data: RetrieveTemplateGroup{
			Total: total,
			TemplateGroups: slice.Map(rts, func(idx int, src domain.TemplateGroup) TemplateGroup {
				return h.toVo(src)
			}),
		},
	}, nil
}

func (h *GroupHandler) toDomain(req CreateTemplateGroupReq) domain.TemplateGroup {
	return domain.TemplateGroup{
		Name: req.Name,
		Icon: req.Icon,
	}
}

func (h *GroupHandler) toVo(req domain.TemplateGroup) TemplateGroup {
	return TemplateGroup{
		Id:   req.Id,
		Name: req.Name,
		Icon: req.Icon,
	}
}
