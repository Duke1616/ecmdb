package web

import (
	"context"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Bunny3th/easy-workflow/workflow/web_api/router"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"strconv"
)

type Handler struct {
	svc       service.Service
	engineSvc engine.Service
}

func NewHandler(svc service.Service, engineSvc engine.Service) *Handler {
	return &Handler{
		svc:       svc,
		engineSvc: engineSvc,
	}
}

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	router.NewRouter(server, "/api/process", false, "")

	g := server.Group("/api/order")
	g.POST("/create", ginx.WrapBody[CreateOrderReq](h.CreateOrder))
	g.POST("/detail/process_inst_id", ginx.WrapBody[DetailProcessInstIdReq](h.Detail))
	g.POST("/todo", ginx.WrapBody[Todo](h.Todo))
	g.POST("/task/record", ginx.WrapBody[RecordTaskReq](h.TaskRecord))
	g.POST("/history", ginx.WrapBody[HistoryReq](h.History))
	g.POST("/start/user", ginx.WrapBody[StartUser](h.StartUser))
	g.POST("/pass", ginx.WrapBody[PassOrderReq](h.Pass))
	g.POST("/reject", ginx.WrapBody[RejectOrderReq](h.Reject))
	g.POST("/list", ginx.WrapBody[Todo](h.Todo))
}

func (h *Handler) CreateOrder(ctx *gin.Context, req CreateOrderReq) (ginx.Result, error) {
	sess, err := session.Get(&gctx.Context{Context: ctx})
	if err != nil {
		return systemErrorResult, fmt.Errorf("获取 Session 失败, %w", err)
	}

	req.CreateBy = strconv.FormatInt(sess.Claims().Uid, 10)
	err = h.svc.CreateOrder(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, fmt.Errorf("创建工单失败, %w", err)
	}

	return ginx.Result{
		Data: "",
		Msg:  "创建工单成功",
	}, nil
}

func (h *Handler) Todo(ctx *gin.Context, req Todo) (ginx.Result, error) {
	// 校验传递参数
	validate := validator.New()
	err := validate.Struct(req)
	if err != nil {
		return validateErrorResult, fmt.Errorf("参数传递错误：%w", err)
	}

	// 查询待处理工单
	instances, total, err := h.engineSvc.ListTodo(ctx, req.UserId, req.ProcessName, req.SortByAsc, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	// 数据处理
	orders, err := h.toVoEngineOrder(ctx, instances)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Data: RetrieveOrders{
			Total: total,
			Tasks: orders,
		},
		Msg: "查看待办工单列表成功",
	}, err
}

func (h *Handler) History(ctx *gin.Context, req HistoryReq) (ginx.Result, error) {
	os, _, err := h.svc.ListHistoryOrder(ctx, req.UserId, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	instIds := slice.Map(os, func(idx int, src domain.Order) int {
		return src.Process.InstanceId
	})

	return ginx.Result{
		Data: instIds,
	}, nil
}

func (h *Handler) Pass(ctx *gin.Context, req PassOrderReq) (ginx.Result, error) {
	err := h.engineSvc.Pass(ctx, req.TaskId, req.Comment)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "同意审批",
		Data: nil,
	}, nil
}

func (h *Handler) Reject(ctx *gin.Context, req RejectOrderReq) (ginx.Result, error) {
	err := h.engineSvc.Reject(ctx, req.TaskId, req.Comment)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "驳回审批",
		Data: nil,
	}, nil
}

func (h *Handler) StartUser(ctx *gin.Context, req StartUser) (ginx.Result, error) {
	instances, total, err := h.engineSvc.ListByStartUser(ctx, req.UserId, req.ProcessName, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	// 数据处理
	orders, err := h.toVoEngineOrder(ctx, instances)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveOrders{
			Total: total,
			Tasks: orders,
		},
		Msg: "查看我的工单列表成功",
	}, err
}

func (h *Handler) Detail(ctx *gin.Context, req DetailProcessInstIdReq) (ginx.Result, error) {
	order, err := h.svc.DetailByProcessInstId(ctx, req.ProcessInstanceId)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: h.toVoOrder(order),
	}, nil
}

func (h *Handler) TaskRecord(ctx *gin.Context, req RecordTaskReq) (ginx.Result, error) {
	ts, total, err := h.engineSvc.TaskRecord(ctx, req.ProcessInstId, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	records := slice.Map(ts, func(idx int, src model.Task) TaskRecord {
		return h.toVoTaskRecords(src)
	})

	return ginx.Result{
		Data: RetrieveTaskRecords{
			TaskRecords: records,
			Total:       total,
		},
	}, nil
}

func (h *Handler) toDomain(req CreateOrderReq) domain.Order {
	return domain.Order{
		CreateBy:   req.CreateBy,
		TemplateId: req.TemplateId,
		WorkflowId: req.WorkflowId,
		Data:       req.Data,
		Status:     domain.START,
	}
}

func (h *Handler) toVoOrder(req domain.Order) Order {
	return Order{
		Data: req.Data,
	}
}

func (h *Handler) toVoTaskRecords(req model.Task) TaskRecord {
	return TaskRecord{
		Nodename:     req.NodeName,
		ApprovedBy:   req.UserID,
		IsCosigned:   req.IsCosigned,
		Status:       req.Status,
		Comment:      req.Comment,
		IsFinished:   req.IsFinished,
		FinishedTime: req.FinishedTime,
	}
}

func (h *Handler) toVoEngineOrder(ctx context.Context, instances []engine.Instance) ([]Order, error) {
	procInstIds := slice.Map(instances, func(idx int, src engine.Instance) int {
		return src.ProcInstID
	})

	os, err := h.svc.ListOrderByProcessInstanceIds(ctx, procInstIds)
	slice.ToMap(os, func(element domain.Order) int64 {
		return element.Id
	})

	m := slice.ToMap(os, func(element domain.Order) int {
		return element.Process.InstanceId
	})

	// 数据转换为前端可用
	return slice.Map(instances, func(idx int, src engine.Instance) Order {
		val, _ := m[src.ProcInstID]
		return Order{
			TaskId:             src.TaskID,
			ProcessInstanceId:  src.ProcInstID,
			Starter:            src.Starter,
			Title:              src.ProcName,
			CurrentStep:        src.CurrentNodeName,
			ApprovedBy:         src.ApprovedBy,
			ProcInstCreateTime: src.CreateTime,
			TemplateId:         val.TemplateId,
			WorkflowId:         val.WorkflowId,
			Ctime:              val.Ctime,
		}
	}), err
}
