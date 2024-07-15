package web

import (
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"strconv"
)

type Handler struct {
	svc     service.Service
	taskSvc task.Service
}

func NewHandler(svc service.Service, taskSvc task.Service) *Handler {
	return &Handler{
		svc:     svc,
		taskSvc: taskSvc,
	}
}

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/api/order")
	g.POST("/create", ginx.WrapBody[CreateOrderReq](h.CreateOrder))
	g.POST("/todo", ginx.WrapBody[Todo](h.Todo))
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

	// 查询未处理工单
	tasks, err := engine.GetTaskToDoList(req.UserId, req.ProcessName, req.SortByAsc, req.Idx, req.Rows)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: tasks,
		Msg:  "查看待办列表成功",
	}, err
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
