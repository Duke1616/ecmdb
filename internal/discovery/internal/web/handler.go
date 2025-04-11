package web

import (
	"github.com/Duke1616/ecmdb/internal/discovery/internal/domain"
	"github.com/Duke1616/ecmdb/internal/discovery/internal/service"
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
	g := server.Group("/api/discovery")
	g.POST("/create", ginx.WrapBody[CreateDiscoveryReq](h.Create))
	g.POST("/update", ginx.WrapBody[UpdateDiscoveryReq](h.Update))
	g.POST("/delete", ginx.WrapBody[DeleteDiscoveryReq](h.Delete))
	g.POST("/list/by_template_id", ginx.WrapBody[ListByTemplateId](h.ListByTemplateId))
}

func (h *Handler) Create(ctx *gin.Context, req CreateDiscoveryReq) (ginx.Result, error) {
	id, err := h.svc.Create(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "创建成功",
		Data: id,
	}, nil
}

func (h *Handler) Delete(ctx *gin.Context, req DeleteDiscoveryReq) (ginx.Result, error) {
	id, err := h.svc.Delete(ctx, req.Id)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "删除成功",
		Data: id,
	}, nil
}

func (h *Handler) ListByTemplateId(ctx *gin.Context, req ListByTemplateId) (ginx.Result, error) {
	rts, total, err := h.svc.ListByTemplateId(ctx, req.Offset, req.Limit, req.TemplateId)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "根据 模版ID 查询自主发现列表成功",
		Data: RetrieveDiscoveries{
			Total: total,
			Discoveries: slice.Map(rts, func(idx int, src domain.Discovery) Discovery {
				return h.toDiscoveryVo(src)
			}),
		},
	}, nil
}

func (h *Handler) Update(ctx *gin.Context, req UpdateDiscoveryReq) (ginx.Result, error) {
	id, err := h.svc.Update(ctx, h.toUpdateDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "修改成功",
		Data: id,
	}, nil
}

func (h *Handler) toDomain(src CreateDiscoveryReq) domain.Discovery {
	return domain.Discovery{
		Field:      src.Field,
		RunnerId:   src.RunnerId,
		TemplateId: src.TemplateId,
		Value:      src.Value,
	}
}

func (h *Handler) toUpdateDomain(src UpdateDiscoveryReq) domain.Discovery {
	return domain.Discovery{
		Id:       src.Id,
		Field:    src.Field,
		RunnerId: src.RunnerId,
		Value:    src.Value,
	}
}

func (h *Handler) toDiscoveryVo(src domain.Discovery) Discovery {
	return Discovery{
		Id:         src.Id,
		Field:      src.Field,
		RunnerId:   src.RunnerId,
		TemplateId: src.TemplateId,
		Value:      src.Value,
	}
}
