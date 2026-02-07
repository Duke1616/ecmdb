package web

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/chromedp/chromedp"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	svc              service.Service
	userSvc          user.Service
	processEngineSvc service.ProcessEngine
	engineSvc        engine.Service
	workflowSvc      workflow.Service
}

func NewHandler(svc service.Service, engineSvc engine.Service, processEngineSvc service.ProcessEngine, userSvc user.Service, workflowSvc workflow.Service) *Handler {
	return &Handler{
		svc:              svc,
		userSvc:          userSvc,
		engineSvc:        engineSvc,
		processEngineSvc: processEngineSvc,
		workflowSvc:      workflowSvc,
	}
}

func (h *Handler) PublicRoute(server *gin.Engine) {
	g := server.Group("/api/order")
	g.POST("/progress", ginx.WrapBody[ProgressReq](h.Progress))
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
	g.POST("/task/form_config", ginx.WrapBody[TaskFormConfigReq](h.GetTaskFormConfig))
}

func (h *Handler) GetTaskFormConfig(ctx *gin.Context, req TaskFormConfigReq) (ginx.Result, error) {
	// 获取当前节点的信息
	info, err := h.engineSvc.TaskInfo(ctx, req.TaskId)
	if err != nil {
		return systemErrorResult, err
	}

	// 查看流程引擎信息
	wf, err := h.workflowSvc.Find(ctx, req.WorkflowId)
	if err != nil {
		return systemErrorResult, err
	}

	nodes, err := easyflow.ParseNodes(wf.FlowData.Nodes)
	if err != nil {
		return systemErrorResult, err
	}

	for _, node := range nodes {
		if node.ID != info.NodeID {
			continue
		}

		property, err1 := easyflow.ToNodeProperty[easyflow.UserProperty](node)
		if err1 != nil {
			return systemErrorResult, err1
		}

		return ginx.Result{
			Data: property.Fields,
			Msg:  "获取任务表单配置成功",
		}, nil
	}

	return ginx.Result{
		Data: []easyflow.Field{},
		Msg:  "未找到对应任务配置",
	}, nil
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
	err = h.processEngineSvc.Revoke(ctx, req.InstanceId, u.Username, req.Force)
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
	if err = h.processEngineSvc.Pass(ctx, req.TaskId, req.Comment, req.ExtraData); err != nil {
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

	err = h.processEngineSvc.Reject(ctx, req.TaskId, req.Comment)
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

	// 1. 提取并去重 UserID
	userIDs := slice.Map(ts, func(idx int, src model.Task) string {
		return src.UserID
	})
	// 简单去重
	uniqueUserIDs := make([]string, 0, len(userIDs))
	for _, uid := range userIDs {
		if !slice.Contains(uniqueUserIDs, uid) {
			uniqueUserIDs = append(uniqueUserIDs, uid)
		}
	}

	// 2. 获取用户 Map (允许部分失败或降级)
	uMap, err := h.getUserMap(ctx, uniqueUserIDs)
	if err != nil {
		// 记录日志但不阻断，前端展示 ID 即可
		// h.l.Warn("Failed to get user map", elog.FieldErr(err))
		uMap = make(map[string]string)
	}

	// 3. 批量获取任务快照数据
	taskIds := slice.Map(ts, func(idx int, src model.Task) int {
		return src.TaskID
	})
	taskDataMap, err := h.svc.FindTaskFormsBatch(ctx, taskIds)
	if err != nil {
		// h.l.Warn("获取任务快照失败", elog.FieldErr(err))
		taskDataMap = make(map[int][]domain.FormValue)
	}

	// 4. 组装返回结果
	records := slice.Map(ts, func(idx int, src model.Task) TaskRecord {
		userName := uMap[src.UserID]
		if userName == "" {
			userName = src.UserID
		}

		return TaskRecord{
			Nodename:     src.NodeName,
			ApprovedBy:   userName,
			IsCosigned:   src.IsCosigned,
			Status:       src.Status,
			Comment:      src.Comment,
			IsFinished:   src.IsFinished,
			FinishedTime: src.FinishedTime,
			FormValues: slice.Map(taskDataMap[src.TaskID], func(idx int, src domain.FormValue) FormValue {
				return FormValue{
					Name:  src.Name,
					Key:   src.Key,
					Type:  src.Type,
					Value: src.Value,
				}
			}),
		}
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
	// 顶层 context，不设超时
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Headless,
		chromedp.NoSandbox,
		chromedp.NoFirstRun,
		chromedp.DisableGPU,
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		//chromedp.Flag("single-process", true),
		chromedp.Flag("no-zygote", true),
		chromedp.Flag("font-render-hinting", "none"),
		chromedp.Flag("force-color-profile", "srgb"),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancelBrowser()

	taskCtx, cancelTask := context.WithTimeout(browserCtx, 30*time.Second)
	defer cancelTask()

	// 存储截图的 buffer
	var buf []byte

	// 定义注入的数据
	injectData := `window.__DATA__ = {
		nodes: [{id: "1", type: "rect", x: 100, y: 100, text: "哈哈哈"}],
		edges: []
	};`

	err := chromedp.Run(taskCtx,
		chromedp.EmulateViewport(1920, 1080, chromedp.EmulateScale(1)),
		chromedp.Navigate(req.TargetUrl),
		chromedp.WaitReady("body"),

		// 注入数据
		chromedp.Evaluate(injectData, nil),

		// 等待 LogicFlow 容器可见
		chromedp.WaitVisible("#LF-preview", chromedp.ByID),

		// 等待前端设置的 data-rendered 标志
		chromedp.WaitVisible(`#LF-preview[data-rendered="true"]`, chromedp.ByQuery),

		// 再等 300ms，防止动画残影
		chromedp.Sleep(300*time.Millisecond),

		// 截取容器截图（比全屏更精准）
		chromedp.Screenshot("#LF-preview", &buf, chromedp.NodeVisible, chromedp.ByID),
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
