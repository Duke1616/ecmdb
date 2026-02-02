package web

import (
	"encoding/json"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/gotomicro/ego/core/elog"

	"github.com/Duke1616/ecmdb/internal/workflow/internal/domain"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/service"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"

	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc       service.Service
	engineSvc engine.Service
	logger    *elog.Component
}

func NewHandler(svc service.Service, engineSvc engine.Service) *Handler {
	return &Handler{
		svc:       svc,
		engineSvc: engineSvc,
		logger:    elog.DefaultLogger.With(elog.FieldComponentName("WorkflowHandler")),
	}
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/workflow")
	g.POST("/create", ginx.WrapBody[CreateReq](h.Create))
	g.POST("/list", ginx.WrapBody[ListReq](h.List))
	g.POST("/update", ginx.WrapBody[UpdateReq](h.Update))
	g.POST("/delete", ginx.WrapBody[DeleteReq](h.Delete))
	g.POST("/deploy", ginx.WrapBody[DeployReq](h.Deploy))

	// 根据关键字搜索流程
	g.POST("/list/by_keyword", ginx.WrapBody[ByKeywordReq](h.ByKeyword))

	// 工单流程图
	g.POST("/graph", ginx.WrapBody[OrderGraphReq](h.FindOrderGraph))
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

func (h *Handler) ByKeyword(ctx *gin.Context, req ByKeywordReq) (ginx.Result, error) {
	ws, total, err := h.svc.FindByKeyword(ctx, req.Keyword, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "根据关键字搜索流程成功",
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

func (h *Handler) FindOrderGraph(ctx *gin.Context, req OrderGraphReq) (ginx.Result, error) {
	// 1. 获取流程实例详情，拿到对应的版本号
	inst, err := h.engineSvc.GetInstanceByID(ctx, req.ProcessInstanceId)
	if err != nil {
		return systemErrorResult, err
	}

	// 2. 获取历史快照 (Version-Aware)
	flow, err := h.svc.FindInstanceFlow(ctx, req.Id, inst.ProcID, inst.ProcVersion)
	if err != nil {
		return ginx.Result{}, err
	}

	// 4. 获取点亮的边
	edgeMap, err := h.engineSvc.GetTraversedEdges(ctx, req.ProcessInstanceId, flow.ProcessId, req.Status)
	if err != nil {
		return systemErrorResult, err
	}

	// 5. 点亮边逻辑
	edgesJSON, _ := json.Marshal(flow.FlowData.Edges)
	var edges []easyflow.Edge
	if err = json.Unmarshal(edgesJSON, &edges); err != nil {
		return systemErrorResult, err
	}

	for i, edge := range edges {
		targets, ok := edgeMap[edge.SourceNodeId]
		if !ok {
			continue
		}

		if slice.Contains(targets, edge.TargetNodeId) {
			properties, ok := edge.Properties.(map[string]interface{})
			if !ok {
				properties = make(map[string]interface{})
			}
			properties["is_pass"] = true
			edges[i].Properties = properties
		}
	}

	// 将处理后的 edges 转回 map 结构
	var newEdges []map[string]interface{}
	newEdgesJSON, _ := json.Marshal(edges)
	_ = json.Unmarshal(newEdgesJSON, &newEdges)
	flow.FlowData.Edges = newEdges

	return ginx.Result{
		Data: RetrieveOrderGraph{
			Workflow: h.toWorkflowVo(flow),
		},
	}, nil
}

func (h *Handler) toDomain(req CreateReq) domain.Workflow {
	return domain.Workflow{
		FlowData: domain.LogicFlow{
			Edges: req.FlowData.Edges,
			Nodes: req.FlowData.Nodes,
		},
		Name:         req.Name,
		Desc:         req.Desc,
		Icon:         req.Icon,
		Owner:        req.Owner,
		IsNotify:     req.IsNotify,
		NotifyMethod: domain.NotifyMethod(req.NotifyMethod),
		TemplateId:   req.TemplateId,
	}
}

func (h *Handler) toUpdateDomain(req UpdateReq) domain.Workflow {
	return domain.Workflow{
		Id:           req.Id,
		Name:         req.Name,
		Desc:         req.Desc,
		Owner:        req.Owner,
		IsNotify:     req.IsNotify,
		NotifyMethod: domain.NotifyMethod(req.NotifyMethod),
		FlowData: domain.LogicFlow{
			Edges: req.FlowData.Edges,
			Nodes: req.FlowData.Nodes,
		},
	}
}

func (h *Handler) toWorkflowVo(req domain.Workflow) Workflow {
	return Workflow{
		Id:           req.Id,
		Name:         req.Name,
		Desc:         req.Desc,
		Icon:         req.Icon,
		Owner:        req.Owner,
		IsNotify:     req.IsNotify,
		NotifyMethod: req.NotifyMethod.ToUint8(),
		TemplateId:   req.TemplateId,
		FlowData: LogicFlow{
			Nodes: req.FlowData.Nodes,
			Edges: req.FlowData.Edges,
		},
	}
}
