package web

import (
	"encoding/json"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
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
	g := server.Group("/api/task")
	g.POST("/list", ginx.WrapBody[ListTaskReq](h.ListTask))
	g.POST("/update/args", ginx.WrapBody[UpdateArgsReq](h.UpdateArgs))
	g.POST("/update/variables", ginx.WrapBody[UpdateVariablesReq](h.UpdateVariableReq))
	g.POST("/retry", ginx.WrapBody[RetryReq](h.Retry))
}

func (h *Handler) ListTask(ctx *gin.Context, req ListTaskReq) (ginx.Result, error) {
	ws, total, err := h.svc.ListTask(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "查询 task 列表成功",
		Data: RetrieveTasks{
			Total: total,
			Tasks: slice.Map(ws, func(idx int, src domain.Task) Task {
				return h.toTaskVo(src)
			}),
		},
	}, nil
}

func (h *Handler) UpdateArgs(ctx *gin.Context, req UpdateArgsReq) (ginx.Result, error) {
	count, err := h.svc.UpdateArgs(ctx, req.Id, req.Args)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
		Msg:  "修改Args成功",
	}, nil
}

func (h *Handler) UpdateVariableReq(ctx *gin.Context, req UpdateVariablesReq) (ginx.Result, error) {
	fmt.Print(req.Variables)
	count, err := h.svc.UpdateVariables(ctx, req.Id, req.Variables)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: count,
		Msg:  "修改Args成功",
	}, nil
}

func (h *Handler) Retry(ctx *gin.Context, req RetryReq) (ginx.Result, error) {
	err := h.svc.RetryTask(ctx, req.Id)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "重试任务成功",
	}, nil
}

func (h *Handler) toTaskVo(req domain.Task) Task {
	args, _ := json.Marshal(req.Args)
	return Task{
		Id:          req.Id,
		OrderId:     req.OrderId,
		Language:    req.Language,
		Code:        req.Code,
		WorkerName:  req.WorkerName,
		CodebookUid: req.CodebookUid,
		Status:      Status(req.Status),
		Result:      req.Result,
		Args:        string(args),
		Variables:   req.Variables,
	}
}
