package web

import (
	"encoding/json"
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

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/task")
	g.POST("/list", ginx.WrapBody[ListTaskReq](h.ListTask))
	g.POST("/update/args", ginx.WrapBody[UpdateArgsReq](h.UpdateArgs))
	g.POST("/update/variables", ginx.WrapBody[UpdateVariablesReq](h.UpdateVariableReq))
	g.POST("/retry", ginx.WrapBody[RetryReq](h.Retry))
	g.POST("/success", ginx.WrapBody[UpdateStatusToSuccessReq](h.UpdateStatusToSuccess))
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

// UpdateStatusToSuccess 当自动化任务卡住, 确实无法完成的情况为了防止工单无法结束、手动修改状态为完成
func (h *Handler) UpdateStatusToSuccess(ctx *gin.Context, req UpdateStatusToSuccessReq) (ginx.Result, error) {
	count, err := h.svc.UpdateTaskStatus(ctx, domain.TaskResult{
		Id:              req.Id,
		TriggerPosition: "手动修改状态为成功",
		WantResult:      "",
		Result:          "",
		Status:          domain.SUCCESS,
	})

	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: count,
		Msg:  "消息状态修改为成功",
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
	variables, err := h.toVariablesDomain(req.Variables)
	if err != nil {
		return systemErrorResult, err
	}

	count, err := h.svc.UpdateVariables(ctx, req.Id, variables)
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
		Id:           req.Id,
		OrderId:      req.OrderId,
		Language:     req.Language,
		Code:         req.Code,
		WorkerName:   req.WorkerName,
		CodebookUid:  req.CodebookUid,
		CodebookName: req.CodebookName,
		Status:       Status(req.Status),
		Result:       req.Result,
		Args:         string(args),
		Variables:    desensitization(req.Variables),
	}
}

func desensitization(req []domain.Variables) string {
	variablesJson := slice.Map(req, func(idx int, src domain.Variables) Variables {
		if src.Secret {
			return Variables{
				Key:    src.Key,
				Value:  "********",
				Secret: src.Secret,
			}
		}

		return Variables{
			Key:    src.Key,
			Value:  src.Value,
			Secret: src.Secret,
		}
	})

	vars, _ := json.Marshal(variablesJson)
	return string(vars)
}

func (h *Handler) toVariablesDomain(variables string) ([]domain.Variables, error) {
	var vars []Variables
	err := json.Unmarshal([]byte(variables), &vars)
	if err != nil {
		return nil, err
	}

	return slice.Map(vars, func(idx int, src Variables) domain.Variables {
		return domain.Variables{
			Key:    src.Key,
			Value:  src.Value,
			Secret: src.Secret,
		}
	}), nil
}
