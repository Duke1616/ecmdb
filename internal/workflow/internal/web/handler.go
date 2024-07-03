package web

import (
	"github.com/Duke1616/ecmdb/internal/workflow/internal/domain"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/service"
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
	g := server.Group("/api/workflow")
	g.POST("/create", ginx.WrapBody[CreateReq](h.Create))
	g.POST("/list", ginx.WrapBody[ListReq](h.List))
}

func (h *Handler) Create(ctx *gin.Context, req CreateReq) (ginx.Result, error) {
	t, err := h.svc.Create(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: t,
	}, nil
}

func (h *Handler) List(ctx *gin.Context, req ListReq) (ginx.Result, error) {
	ws, total, err := h.svc.List(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "查询流程模版列表成功",
		Data: RetrieveWorkflows{
			Total: total,
			Workflows: slice.Map(ws, func(idx int, src domain.Workflow) Workflow {
				return h.toWorkflowVo(src)
			}),
		},
	}, nil
}

func (h *Handler) toDomain(req CreateReq) domain.Workflow {
	return domain.Workflow{
		FlowData:   req.FlowData,
		Name:       req.Name,
		Desc:       req.Desc,
		Icon:       req.Icon,
		Owner:      req.Owner,
		TemplateId: req.TemplateId,
	}
}

func (h *Handler) toWorkflowVo(req domain.Workflow) Workflow {
	return Workflow{
		Id:         req.Id,
		Name:       req.Name,
		Desc:       req.Desc,
		Icon:       req.Icon,
		Owner:      req.Owner,
		TemplateId: req.TemplateId,
		FlowData:   req.FlowData,
	}
}
