package web

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	executorv1 "github.com/Duke1616/ecmdb/api/proto/gen/etask/executor/v1"
	"github.com/Duke1616/ecmdb/internal/task/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc         service.Service
	executorSvc executorv1.TaskExecutionServiceClient
}

func NewHandler(svc service.Service, executorSvc executorv1.TaskExecutionServiceClient) *Handler {
	return &Handler{
		svc:         svc,
		executorSvc: executorSvc,
	}
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/task")
	g.POST("/list", ginx.WrapBody[ListTaskReq](h.ListTask))
	g.POST("/list/by_instance_id", ginx.WrapBody[ListTaskByInstanceIdReq](h.ListTaskByInstanceId))
	g.POST("/update/args", ginx.WrapBody[UpdateArgsReq](h.UpdateArgs))
	g.POST("/update/variables", ginx.WrapBody[UpdateVariablesReq](h.UpdateVariableReq))
	g.POST("/retry", ginx.WrapBody[RetryReq](h.Retry))
	g.POST("/success", ginx.WrapBody[UpdateStatusToSuccessReq](h.UpdateStatusToSuccess))
	g.GET("/logs/:task_id", ginx.Wrap(h.Logs))
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

func (h *Handler) ListTaskByInstanceId(ctx *gin.Context, req ListTaskByInstanceIdReq) (ginx.Result, error) {
	ws, total, err := h.svc.ListTaskByInstanceId(ctx, req.Offset, req.Limit, req.InstanceId)
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
		TriggerPosition: domain.TriggerPositionManualSuccess.ToString(),
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

func (h *Handler) Logs(ctx *gin.Context) (ginx.Result, error) {
	idStr := ctx.Param("task_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return systemErrorResult, err
	}

	tInfo, err := h.svc.Detail(ctx, id)
	if err != nil {
		return systemErrorResult, err
	}

	// 如果是 KAFKA 模式直接返回 result 数据
	if tInfo.Kind == domain.KAFKA {
		return ginx.Result{
			Code: 0,
			Msg:  "获取日志成功",
			Data: tInfo.Result,
		}, nil
	}

	// 如果是 EXECUTE 模式，但还没派发（没有 ExternalId），返回空提示
	if tInfo.ExternalId == "" {
		return ginx.Result{
			Code: 0,
			Msg:  "任务尚未派发或正在队列中，暂无执行日志",
			Data: "任务尚未派发或正在队列中，暂无执行日志",
		}, nil
	}

	externalTaskId, err := strconv.ParseInt(tInfo.ExternalId, 10, 64)
	if err != nil {
		return systemErrorResult, fmt.Errorf("非法 ExternalId: %w", err)
	}

	// 1. 获取该任务的所有执行历史
	executionsResp, err := h.executorSvc.ListTaskExecutions(ctx, &executorv1.ListTaskExecutionsRequest{
		TaskId: externalTaskId,
	})
	if err != nil {
		return systemErrorResult, err
	}

	if len(executionsResp.Executions) == 0 {
		return ginx.Result{
			Code: 0,
			Msg:  "暂无执行记录",
			Data: "暂无执行记录",
		}, nil
	}

	// 2. 取最后一次执行记录（最新的执行记录）
	executionId := executionsResp.Executions[len(executionsResp.Executions)-1].Id

	// 3. 分页拉取（这里先拉取前 1000 条，基本覆盖绝大多数 shell 脚本场景）
	logsResp, err := h.executorSvc.GetExecutionLogs(ctx, &executorv1.GetExecutionLogsRequest{
		ExecutionId: executionId,
		Limit:       1000,
	})
	if err != nil {
		return systemErrorResult, err
	}

	// 4. 聚合日志行
	var sb strings.Builder
	for _, l := range logsResp.Logs {
		sb.WriteString(l.Content)
	}

	return ginx.Result{
		Code: 0,
		Msg:  "获取日志成功",
		Data: sb.String(),
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

	// 定时任务以 ScheduledTime 作为计划开始时间；普通任务则以 Utime 兜底
	scheduledTime := time.UnixMilli(req.Utime).Format("2006-01-02 15:04:05")
	if req.IsTiming {
		scheduledTime = time.UnixMilli(req.ScheduledTime).Format("2006-01-02 15:04:05")
	}

	// 真实的任务执行开始/结束时间存储在 StartTime、EndTime 字段，值为 0 表示尚未执行
	var startTime, endTime string
	if req.StartTime > 0 {
		startTime = time.UnixMilli(req.StartTime).Format("2006-01-02 15:04:05")
	}
	if req.EndTime > 0 {
		endTime = time.UnixMilli(req.EndTime).Format("2006-01-02 15:04:05")
	}

	taskVO := Task{
		Id:              req.Id,
		OrderId:         req.OrderId,
		Language:        req.Language,
		Code:            req.Code,
		Kind:            string(req.Kind),
		CodebookUid:     req.CodebookUid,
		CodebookName:    req.CodebookName,
		Target:          req.Target,
		Handler:         req.Handler,
		Status:          Status(req.Status),
		Result:          req.Result,
		Args:            string(args),
		IsTiming:        req.IsTiming,
		ScheduledTime:   scheduledTime,
		StartTime:       startTime,
		EndTime:         endTime,
		RetryCount:      req.RetryCount,
		TriggerPosition: req.TriggerPosition,
		Variables:       desensitization(req.Variables),
	}

	return taskVO
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
