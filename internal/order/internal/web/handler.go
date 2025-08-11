package web

import (
	"context"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/chromedp/chromedp"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"io/ioutil"
	"log"
	"time"
)

type Handler struct {
	svc       service.Service
	userSvc   user.Service
	engineSvc engine.Service
}

func NewHandler(svc service.Service, engineSvc engine.Service, userSvc user.Service) *Handler {
	return &Handler{
		svc:       svc,
		userSvc:   userSvc,
		engineSvc: engineSvc,
	}
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	//router.NewRouter(server, "/api/process", false, "")
	g := server.Group("/api/order")
	g.POST("/create", ginx.WrapBody[CreateOrderReq](h.CreateOrder))
	g.POST("/detail/process_inst_id", ginx.WrapBody[DetailProcessInstIdReq](h.Detail))
	g.POST("/task/record", ginx.WrapBody[RecordTaskReq](h.TaskRecord))
	g.POST("/todo", ginx.WrapBody[Todo](h.TodoAll))
	g.POST("/todo/user", ginx.WrapBody[Todo](h.TodoByUser))
	g.POST("/history", ginx.WrapBody[HistoryReq](h.History))
	g.POST("/start/user", ginx.WrapBody[StartUserReq](h.StartUser))
	g.POST("/pass", ginx.WrapBody[PassOrderReq](h.Pass))
	g.POST("/reject", ginx.WrapBody[RejectOrderReq](h.Reject))
	g.POST("/revoke", ginx.WrapBody[RevokeOrderReq](h.Revoke))
	g.POST("/progress", ginx.WrapBody[ProgressReq](h.Progress))
}

func (h *Handler) CreateOrder(ctx *gin.Context, req CreateOrderReq) (ginx.Result, error) {
	if req.CreateBy == "" {
		u, err := h.getSessUser(ctx)
		if err != nil {
			return systemErrorResult, err
		}
		req.CreateBy = u.Username
	}

	err := h.svc.CreateOrder(ctx, h.toDomain(req))
	if err != nil {
		return systemErrorResult, fmt.Errorf("创建工单失败, %w", err)
	}

	return ginx.Result{
		Data: "",
		Msg:  "创建工单成功",
	}, nil
}

// TodoAll 全部待办
func (h *Handler) TodoAll(ctx *gin.Context, req Todo) (ginx.Result, error) {
	// 校验传递参数
	validate := validator.New()
	err := validate.Struct(req)
	if err != nil {
		return validateErrorResult, fmt.Errorf("参数传递错误：%w", err)
	}

	// 查询待处理工单
	instances, total, err := h.engineSvc.ListTodoTasks(ctx, req.UserId, req.ProcessName, req.SortByAsc,
		req.Offset, req.Limit)
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
		Msg: "查看待办工单列表成功",
	}, err
}

// TodoByUser 我的待办
func (h *Handler) TodoByUser(ctx *gin.Context, req Todo) (ginx.Result, error) {
	// 校验传递参数
	validate := validator.New()
	err := validate.Struct(req)
	if err != nil {
		return validateErrorResult, fmt.Errorf("参数传递错误：%w", err)
	}

	// 查询用户信息 - 为了统一存储为 username
	u, err := h.getSessUser(ctx)
	if err != nil {
		return systemErrorResult, err
	}

	// 查询待处理工单
	instances, total, err := h.engineSvc.ListTodoTasks(ctx, u.Username, req.ProcessName, req.SortByAsc,
		req.Offset, req.Limit)
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

func (h *Handler) Revoke(ctx *gin.Context, req RevokeOrderReq) (ginx.Result, error) {
	u, err := h.getSessUser(ctx)
	if err != nil {
		return systemErrorResult, err
	}

	// 撤销流程工单
	err = h.engineSvc.Revoke(ctx, req.InstanceId, u.Username, req.Force)
	if err != nil {
		return systemErrorResult, err
	}

	// 修改状态
	err = h.svc.UpdateStatusByInstanceId(ctx, req.InstanceId, domain.WITHDRAW.ToUint8())
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "撤销工单成功",
		Data: true,
	}, nil
}

// History 历史工单
func (h *Handler) History(ctx *gin.Context, req HistoryReq) (ginx.Result, error) {
	os, total, err := h.svc.ListHistoryOrder(ctx, req.UserId, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}
	uniqueMap := make(map[string]bool)
	uns := slice.FilterMap(os, func(idx int, src domain.Order) (string, bool) {
		if !uniqueMap[src.CreateBy] {
			uniqueMap[src.CreateBy] = true
			return src.CreateBy, true
		}

		return src.CreateBy, false
	})

	uMap, err := h.getUserMap(ctx, uns)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveOrders{
			Total: total,
			Tasks: slice.Map(os, func(idx int, src domain.Order) Order {
				starter, ok := uMap[src.CreateBy]
				if !ok {
					starter = src.CreateBy
				}

				return Order{
					Id:                src.Id,
					TemplateId:        src.TemplateId,
					Starter:           starter,
					Status:            src.Status.ToUint8(),
					Provide:           src.Provide.ToUint8(),
					ProcessInstanceId: src.Process.InstanceId,
					WorkflowId:        src.WorkflowId,
					Ctime:             time.Unix(src.Ctime/1000, 0).Format("2006-01-02 15:04:05"),
					Wtime:             time.Unix(src.Wtime/1000, 0).Format("2006-01-02 15:04:05"),
					Data:              src.Data,
				}
			}),
		},
	}, nil
}

func (h *Handler) Pass(ctx *gin.Context, req PassOrderReq) (ginx.Result, error) {
	err := h.verifyUser(ctx, req.TaskId)
	if err != nil {
		return systemErrorResult, err
	}

	err = h.engineSvc.Pass(ctx, req.TaskId, req.Comment)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "同意审批",
		Data: nil,
	}, nil
}

func (h *Handler) Reject(ctx *gin.Context, req RejectOrderReq) (ginx.Result, error) {
	err := h.verifyUser(ctx, req.TaskId)
	if err != nil {
		return systemErrorResult, err
	}

	err = h.engineSvc.Reject(ctx, req.TaskId, req.Comment)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "驳回审批",
		Data: nil,
	}, nil
}

// StartUser 与我相关的工单
func (h *Handler) StartUser(ctx *gin.Context, req StartUserReq) (ginx.Result, error) {
	u, err := h.getSessUser(ctx)
	if err != nil {
		return systemErrorResult, err
	}

	// 查找本地存储，关于我的工单（未完成）
	orders, total, err := h.svc.ListOrdersByUser(ctx, u.Username, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	// 查找出所有的流程ID
	procInstIds := slice.Map(orders, func(idx int, src domain.Order) int {
		return src.Process.InstanceId
	})

	// 查询所有流程等待运行的信息 ( 当前步骤、审批人 ）
	processTasks, err := h.engineSvc.ListPendingStepsOfMyTask(ctx, procInstIds, u.Username)
	if err != nil {
		return systemErrorResult, err
	}

	// 数据处理
	tasks, err := h.toVoEngineOrder(ctx, processTasks)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveOrders{
			Total: total,
			Tasks: tasks,
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

// TaskRecord 任务记录
func (h *Handler) TaskRecord(ctx *gin.Context, req RecordTaskReq) (ginx.Result, error) {
	ts, total, err := h.engineSvc.TaskRecord(ctx, req.ProcessInstId, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	uniqueMap := make(map[string]bool)
	uns := slice.FilterMap(ts, func(idx int, src model.Task) (string, bool) {
		if !uniqueMap[src.UserID] {
			uniqueMap[src.UserID] = true
			return src.UserID, true
		}

		return src.UserID, false
	})

	uMap, err := h.getUserMap(ctx, uns)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: RetrieveTaskRecords{
			TaskRecords: slice.Map(ts, func(idx int, src model.Task) TaskRecord {
				starter, ok := uMap[src.UserID]
				if !ok {
					starter = src.UserID
				}
				return TaskRecord{
					Nodename:     src.NodeName,
					ApprovedBy:   starter,
					IsCosigned:   src.IsCosigned,
					Status:       src.Status,
					Comment:      src.Comment,
					IsFinished:   src.IsFinished,
					FinishedTime: src.FinishedTime,
				}
			}),
			Total: total,
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
		Provide:    domain.SYSTEM,
	}
}

func (h *Handler) toVoOrder(req domain.Order) Order {
	return Order{
		TemplateId:        req.TemplateId,
		Starter:           req.CreateBy,
		ProcessInstanceId: req.Process.InstanceId,
		Provide:           req.Provide.ToUint8(),
		Status:            req.Status.ToUint8(),
		WorkflowId:        req.WorkflowId,
		Ctime:             time.Unix(req.Ctime/1000, 0).Format("2006-01-02 15:04:05"),
		Wtime:             time.Unix(req.Wtime/1000, 0).Format("2006-01-02 15:04:05"),
		Data:              req.Data,
	}
}

// 废弃不使用、交由前端处理
func (h *Handler) toTasks(instances []engine.Instance) map[int][]engine.Instance {
	var tasks map[int][]engine.Instance
	tasks = slice.ToMapV(instances, func(m engine.Instance) (int, []engine.Instance) {
		return m.ProcInstID, slice.FilterMap(instances, func(idx int, src engine.Instance) (engine.Instance, bool) {
			if m.ProcInstID == src.ProcInstID {
				return src, true
			}
			return engine.Instance{}, false
		})
	})

	return tasks
}

// 废弃不使用、交由前端处理
func (h *Handler) toSteps(instances []engine.Instance) []Steps {
	var tempSteps map[string][]string
	tempSteps = slice.ToMapV(instances, func(m engine.Instance) (string, []string) {
		return m.CurrentNodeName, slice.FilterMap(instances, func(idx int, src engine.Instance) (string, bool) {
			if m.CurrentNodeName == src.CurrentNodeName {
				return src.ApprovedBy, true
			}
			return "", false
		})
	})

	var steps []Steps
	for k, v := range tempSteps {
		steps = append(steps, Steps{
			CurrentStep: k,
			ApprovedBy:  v,
		})
	}

	return steps
}

func (h *Handler) Progress(ctx *gin.Context, req ProgressReq) (ginx.Result, error) {
	gCtx, cancel := chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// 设置超时时间
	gCtx, cancel = context.WithTimeout(gCtx, 5*time.Second)
	defer cancel()

	// 存储截图的 buffer
	var buf []byte

	err := chromedp.Run(gCtx,
		chromedp.Navigate(req.TargetUrl),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Page loaded, waiting for LF-preview...")
			return nil
		}),
		chromedp.Evaluate(`window.__DATA__ = {nodes: [{id: "1", type: "rect", x: 100, y: 100, text: "哈哈哈"}], edges: []};`, nil),
		chromedp.WaitVisible("#LF-preview", chromedp.ByID), // 使用 ID 确保精准选择
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("LF-preview is visible, capturing screenshot...")
			return nil
		}),
		chromedp.FullScreenshot(&buf, 2000),
	)

	if err != nil {
		return ginx.Result{}, err
	}

	// 保存截图到文件
	err = ioutil.WriteFile("logicflow.png", buf, 0644)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{}, nil
}
func (h *Handler) toVoEngineOrder(ctx context.Context, instances []engine.Instance) ([]Order, error) {
	if len(instances) == 0 {
		// 没有工单信息
		return nil, nil
	}

	uniqueProcInstIds := make(map[int]bool)
	procInstIds := slice.FilterMap(instances, func(idx int, src engine.Instance) (int, bool) {
		if !uniqueProcInstIds[src.ProcInstID] {
			uniqueProcInstIds[src.ProcInstID] = true
			return src.ProcInstID, true
		}

		return src.ProcInstID, false
	})

	us, err := h.getUsers(ctx, instances)
	if err != nil {
		return nil, err
	}

	os, err := h.svc.ListOrderByProcessInstanceIds(ctx, procInstIds)
	m := slice.ToMap(os, func(element domain.Order) int {
		return element.Process.InstanceId
	})

	// 数据转换为前端可用
	return slice.Map(instances, func(idx int, src engine.Instance) Order {
		val, _ := m[src.ProcInstID]
		starter, ok := us[src.Starter]
		if !ok {
			starter = src.Starter
		}
		approved, ok := us[src.ApprovedBy]
		if !ok {
			approved = src.ApprovedBy
		}
		return Order{
			Id:                 val.Id,
			TaskId:             src.TaskID,
			ProcessInstanceId:  src.ProcInstID,
			Starter:            starter,
			CurrentStep:        src.CurrentNodeName,
			ApprovedBy:         approved,
			ProcInstCreateTime: src.CreateTime,
			Provide:            val.Provide.ToUint8(),
			Status:             val.Status.ToUint8(),
			TemplateId:         val.TemplateId,
			WorkflowId:         val.WorkflowId,
			Ctime:              time.Unix(val.Ctime/1000, 0).Format("2006-01-02 15:04:05"),
		}
	}), err
}

func (h *Handler) getUsers(ctx context.Context, instances []engine.Instance) (map[string]string, error) {
	var uns []string
	uniqueMap := make(map[string]bool)
	approved := slice.FilterMap(instances, func(idx int, src engine.Instance) (string, bool) {
		if !uniqueMap[src.ApprovedBy] {
			uniqueMap[src.ApprovedBy] = true
			return src.ApprovedBy, true
		}

		return src.ApprovedBy, false
	})

	starter := slice.FilterMap(instances, func(idx int, src engine.Instance) (string, bool) {
		if !uniqueMap[src.Starter] {
			uniqueMap[src.Starter] = true
			return src.Starter, true
		}

		return src.Starter, false
	})

	uns = append(uns, approved...)
	uns = append(uns, starter...)

	return h.getUserMap(ctx, uns)
}

// 根据 Sess 获取用户
func (h *Handler) getSessUser(ctx *gin.Context) (user.User, error) {
	// 获取登录用户 sess 获取ID
	sess, err := session.Get(&gctx.Context{Context: ctx})
	if err != nil {
		return user.User{}, fmt.Errorf("获取 Session 失败, %w", err)
	}

	// 查询用户信息 - 为了统一存储为 username
	return h.userSvc.FindById(ctx, sess.Claims().Uid)
}

func (h *Handler) verifyUser(ctx *gin.Context, taskId int) error {
	userInfo, err := h.getSessUser(ctx)
	if err != nil {
		return err
	}

	// 如果不是管理员用户，需要进行验证
	if !isAdmin(userInfo.RoleCodes) {
		tInfo := model.Task{}
		tInfo, err = h.engineSvc.TaskInfo(ctx, taskId)
		if err != nil {
			return err
		}

		if tInfo.UserID != userInfo.Username {
			return fmt.Errorf("无法操作，任务审批用户不一致")
		}
	}

	return nil
}

func isAdmin(roleCodes []string) bool {
	for _, code := range roleCodes {
		if code != "admin" {
			return true
		}
	}

	return false
}

// 获取用户Map映射
func (h *Handler) getUserMap(ctx context.Context, uns []string) (map[string]string, error) {
	us, err := h.userSvc.FindByUsernames(ctx, uns)
	if err != nil {
		return nil, err
	}

	return slice.ToMapV(us, func(element user.User) (string, string) {
		return element.Username, element.DisplayName
	}), nil
}
