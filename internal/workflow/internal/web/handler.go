package web

import (
	"fmt"
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
	g.POST("/update", ginx.WrapBody[UpdateReq](h.Update))
	g.POST("/delete", ginx.WrapBody[DeleteReq](h.Delete))
	g.POST("/deploy", ginx.WrapBody[DeployReq](h.Deploy))
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

// Deploy 发布到流程控制系统
func (h *Handler) Deploy(ctx *gin.Context, req DeployReq) (ginx.Result, error) {
	flow, err := h.svc.Find(ctx, req.Id)
	if err != nil {
		return systemErrorResult, fmt.Errorf("查询节点信息失败")
	}

	err = h.svc.Deploy(ctx, flow)
	if err != nil {
		return systemErrorResult, fmt.Errorf("发布失败")
	}
	return ginx.Result{
		Data: flow,
	}, nil
}

func (h *Handler) Update(ctx *gin.Context, req UpdateReq) (ginx.Result, error) {
	t, err := h.svc.Update(ctx, h.toUpdateDomain(req))

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: t,
	}, nil
}

func (h *Handler) Delete(ctx *gin.Context, req DeleteReq) (ginx.Result, error) {
	count, err := h.svc.Delete(ctx, req.Id)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
	}, nil
}

func (h *Handler) toDomain(req CreateReq) domain.Workflow {
	return domain.Workflow{
		FlowData: domain.LogicFlow{
			Edges: req.FlowData.Edges,
			Nodes: req.FlowData.Nodes,
		},
		Name:       req.Name,
		Desc:       req.Desc,
		Icon:       req.Icon,
		Owner:      req.Owner,
		TemplateId: req.TemplateId,
	}
}

func (h *Handler) toUpdateDomain(req UpdateReq) domain.Workflow {
	return domain.Workflow{
		Id:    req.Id,
		Name:  req.Name,
		Owner: req.Owner,
		FlowData: domain.LogicFlow{
			Edges: req.FlowData.Edges,
			Nodes: req.FlowData.Nodes,
		},
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
		FlowData: LogicFlow{
			Nodes: req.FlowData.Nodes,
			Edges: req.FlowData.Edges,
		},
	}
}
