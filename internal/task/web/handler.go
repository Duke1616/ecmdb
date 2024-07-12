package web

import (
	"encoding/json"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(server *gin.Engine) {
	// easy-workflow 路由注册
	// 因为会有一些定制化的需求，决定自己重写路由
	// router.NewRouter(server, "/api/process", false, "")

	// 本地服务路由注册
	g := server.Group("/api/task")

	g.POST("/start", ginx.WrapBody[StartTaskReq](h.StartTask))
	g.POST("/todo", ginx.WrapBody[TodoListTaskReq](h.Todo))
	g.POST("/pass", ginx.WrapBody[PassTaskReq](h.Pass))
}

func (h *Handler) StartTask(ctx *gin.Context, req StartTaskReq) (ginx.Result, error) {
	VariablesJson, err := json.Marshal(req.Variables)
	if err != nil {
		return systemErrorResult, err
	}

	id, err := engine.InstanceStart(req.ProcessId, req.BusinessId, req.Comment, string(VariablesJson))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
		Msg:  "创建任务成功",
	}, nil
}

func (h *Handler) Todo(ctx *gin.Context, req TodoListTaskReq) (ginx.Result, error) {
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

func (h *Handler) Pass(ctx *gin.Context, req PassTaskReq) (ginx.Result, error) {
	err := engine.TaskPass(req.TaskId, req.Comment, req.VariablesJson, false)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "流程节点通过",
		Data: "",
	}, err
}
