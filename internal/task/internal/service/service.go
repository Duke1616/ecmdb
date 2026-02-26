package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/discovery"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/runner"
	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	// CreateTask 仅仅做任务数据的抢占式登记创建，默认设定为等待状态
	CreateTask(ctx context.Context, processInstId int, nodeId string) (int64, error)

	// StartTask 在节点触发时被调起，会聚合工单、工作流信息、调度 Worker 规则以驱动下层引擎开火
	StartTask(ctx context.Context, processInstId int, nodeId string) error

	// RetryTask 对执行异常失败的节点任务启动安全重试补发机制
	RetryTask(ctx context.Context, id int64) error

	// UpdateTaskStatus 被底层异步执行通道回调更新回调结果
	UpdateTaskStatus(ctx context.Context, req domain.TaskResult) (int64, error)

	// UpdateArgs 动态修改参数上下文环境信息
	UpdateArgs(ctx context.Context, id int64, args map[string]interface{}) (int64, error)

	// UpdateVariables 修改执行环境变量内容（带有敏感字段防篡改规则）
	UpdateVariables(ctx context.Context, id int64, variables []domain.Variables) (int64, error)

	// ListTaskByStatus 用于平台展示过滤指定生命周期下的节点任务
	ListTaskByStatus(ctx context.Context, offset, limit int64, status uint8) ([]domain.Task, int64, error)

	// ListTask 全景展示系统大盘所积累的所有执行节点情况
	ListTask(ctx context.Context, offset, limit int64) ([]domain.Task, int64, error)

	// ListTaskByInstanceId 查看某个流程跑过哪些流水线环节实体
	ListTaskByInstanceId(ctx context.Context, offset, limit int64, instanceId int) ([]domain.Task, int64, error)

	// ListSuccessTasksByUtime 事件中心提取增量处理完毕结果集的调度窗口方法
	ListSuccessTasksByUtime(ctx context.Context, offset, limit int64, utime int64) ([]domain.Task, int64, error)

	// FindTaskResult 供周边探针寻找某固定流转节点输出成效
	FindTaskResult(ctx context.Context, instanceId int, nodeId string) (domain.Task, error)

	// Detail 完整还原某一子执行节点的细节与过程属性
	Detail(ctx context.Context, id int64) (domain.Task, error)

	// MarkTaskAsAutoPassed 在引擎收到通知并向前流转后打点忽略重复消息投递验证
	MarkTaskAsAutoPassed(ctx context.Context, id int64) error
}

type service struct {
	repo         repository.TaskRepository
	logger       *elog.Component
	orderSvc     order.Service
	userSvc      user.Service
	discoverySvc discovery.Service
	engineSvc    engine.Service
	workflowSvc  workflow.Service
	codebookSvc  codebook.Service
	runnerSvc    runner.Service
	execSvc      ExecService
	cronjobSvc   Cronjob
}

func (s *service) ListTaskByInstanceId(ctx context.Context, offset, limit int64, instanceId int) ([]domain.Task, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Task
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListTaskByInstanceId(ctx, offset, limit, instanceId)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalByInstanceId(ctx, instanceId)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) Detail(ctx context.Context, id int64) (domain.Task, error) {
	return s.repo.FindById(ctx, id)
}

func (s *service) FindTaskResult(ctx context.Context, instanceId int, nodeId string) (domain.Task, error) {
	return s.repo.FindTaskResult(ctx, instanceId, nodeId)
}

func (s *service) MarkTaskAsAutoPassed(ctx context.Context, id int64) error {
	return s.repo.MarkTaskAsAutoPassed(ctx, id)
}

func (s *service) ListSuccessTasksByUtime(ctx context.Context, offset, limit int64, utime int64) ([]domain.Task, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Task
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListSuccessTasksByUtime(ctx, offset, limit, utime)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalByUtime(ctx, utime)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) ListTask(ctx context.Context, offset, limit int64) ([]domain.Task, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Task
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListTask(ctx, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx, 0)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) CreateTask(ctx context.Context, processInstId int, nodeId string) (int64, error) {
	taskId, err := s.repo.CreateTask(ctx, domain.Task{
		ProcessInstId:   processInstId,
		TriggerPosition: "任务等待",
		CurrentNodeId:   nodeId,
		Status:          domain.WAITING,
	})

	return taskId, err
}

func NewService(repo repository.TaskRepository, orderSvc order.Service, workflowSvc workflow.Service,
	codebookSvc codebook.Service, runnerSvc runner.Service, cronjobSvc Cronjob, engineSvc engine.Service,
	userSvc user.Service, execSvc ExecService, discoverySvc discovery.Service) Service {
	return &service{
		repo:         repo,
		logger:       elog.DefaultLogger,
		orderSvc:     orderSvc,
		workflowSvc:  workflowSvc,
		codebookSvc:  codebookSvc,
		runnerSvc:    runnerSvc,
		engineSvc:    engineSvc,
		userSvc:      userSvc,
		execSvc:      execSvc,
		cronjobSvc:   cronjobSvc,
		discoverySvc: discoverySvc,
	}
}

func (s *service) StartTask(ctx context.Context, processInstId int, nodeId string) error {
	// 避免并发双写和流程空洞：原子化创建或查找当前流程的任务挂载点
	task, err := s.repo.FindOrCreate(ctx, domain.Task{
		ProcessInstId:   processInstId,
		CurrentNodeId:   nodeId,
		TriggerPosition: "准备启动节点",
		Status:          domain.SCHEDULE,
	})

	if err != nil {
		s.logger.Error("获取任务工作锚点失败",
			elog.FieldErr(err),
			elog.Any("流程实例ID", processInstId),
			elog.Any("当前节点ID", nodeId),
		)
		return err
	}

	// 驱动其正式跑起来
	return s.process(ctx, task)
}

func (s *service) RetryTask(ctx context.Context, id int64) error {
	task, err := s.repo.FindById(ctx, id)

	if err != nil {
		elog.Error("重试任务失败",
			elog.Any("错误信息", err),
			elog.Any("任务ID", id),
		)
		return err
	}

	// 证明这个任务是失败的，应该重新获取数据
	if task.OrderId == 0 {
		return s.process(ctx, task)
	}

	return s.retry(ctx, task)
}

func (s *service) UpdateArgs(ctx context.Context, id int64, args map[string]interface{}) (int64, error) {
	return s.repo.UpdateArgs(ctx, id, args)
}

func (s *service) UpdateVariables(ctx context.Context, id int64, variables []domain.Variables) (int64, error) {
	task, err := s.repo.FindById(ctx, id)
	if err != nil {
		return 0, err
	}

	mapKey := slice.ToMap(task.Variables, func(element domain.Variables) string {
		return element.Key
	})

	variables = slice.Map(variables, func(idx int, src domain.Variables) domain.Variables {
		val, ok := mapKey[src.Key]
		if ok {
			// 拒绝对于敏感信息的修改
			if val.Secret {
				return domain.Variables{
					Key:    src.Key,
					Value:  val.Value,
					Secret: val.Secret,
				}
			} else {
				return domain.Variables{
					Key:    src.Key,
					Value:  src.Value,
					Secret: val.Secret,
				}
			}
		}

		// 如果修改了 key 的话，那我也没有什么办法了
		return domain.Variables{
			Key:    src.Key,
			Value:  src.Value,
			Secret: src.Secret,
		}
	})

	return s.repo.UpdateVariables(ctx, id, variables)
}

func (s *service) ListTaskByStatus(ctx context.Context, offset, limit int64, status uint8) ([]domain.Task, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Task
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListTaskByStatus(ctx, offset, limit, status)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx, status)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) UpdateTaskStatus(ctx context.Context, req domain.TaskResult) (int64, error) {
	return s.repo.UpdateTaskStatus(ctx, req)
}

func (s *service) retry(ctx context.Context, task domain.Task) error {
	_, _ = s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
		Id:              task.Id,
		TriggerPosition: "重试任务",
		Status:          domain.RETRY,
	})

	return s.execSvc.Execute(ctx, task)
}

// taskProcessContext 封装了流转任务执行时所需的完整依赖链路元数据
type taskProcessContext struct {
	order      order.Order
	inst       engine.Instance
	flow       workflow.Workflow
	automation easyflow.AutomationProperty
	runner     runner.Runner
	codebook   codebook.Codebook
}

func (s *service) process(ctx context.Context, task domain.Task) error {
	// 1. 获取并聚合所有前置运行依赖的上下文（工单、流程快照、规则、被调度的执行节点等）
	pCtx, err := s.buildTaskProcessContext(ctx, task)
	if err != nil {
		return err // 内部已处理状态变迁与打点，直接冒泡报错
	}

	// 2. 合成包含用户凭单上下文等在运行时的最终参数
	args, err := s.assembleRuntimeArgs(ctx, pCtx.order)
	if err != nil {
		// 装配参数属于偏业务态的软异常或数据缺失，通常暂不强退节点状态（保持原逻辑）
		return err
	}

	// 3. 计算节点的触发模式（定时阻塞调度，还是立即开火）
	status, timing := s.determineTaskStatus(pCtx.automation, pCtx.order.Data)

	// 4. 将以上所有快照合并更新到该任务条目的实体存盘上
	taskUpdate := s.prepareTaskUpdate(
		pCtx.order, task, pCtx.flow, pCtx.codebook,
		pCtx.runner, args, status, timing, pCtx.automation,
	)

	if _, err = s.repo.UpdateTask(ctx, taskUpdate); err != nil {
		return err // 存盘失败抛出
	}

	// 5. 最终路由分配
	return s.dispatchTask(ctx, taskUpdate, pCtx.automation)
}

func (s *service) dispatchTask(ctx context.Context, task domain.Task, automation easyflow.AutomationProperty) error {
	if automation.IsTiming {
		go func() {
			_ = s.cronjobSvc.Create(ctx, task)
		}()
		return nil
	}

	return s.execSvc.Execute(ctx, task)
}

func (s *service) buildTaskProcessContext(ctx context.Context, task domain.Task) (*taskProcessContext, error) {
	// 1. 获取工单信息
	orderResp, err := s.orderSvc.DetailByProcessInstId(ctx, task.ProcessInstId)
	if err != nil {
		return nil, s.handleTaskError(ctx, task.Id, "获取工单失败", domain.FAILED, err)
	}

	// 2. 获取流程实例详情，拿到对应的版本号
	inst, err := s.engineSvc.GetInstanceByID(ctx, task.ProcessInstId)
	if err != nil {
		return nil, s.handleTaskError(ctx, task.Id, "获取流程实例失败", domain.FAILED, err)
	}

	// 3. 尝试获取历史快照
	flow, err := s.workflowSvc.FindInstanceFlow(ctx, orderResp.WorkflowId, inst.ProcID, inst.ProcVersion)
	if err != nil {
		return nil, s.handleTaskError(ctx, task.Id, "获取流程信息失败", domain.FAILED, err)
	}

	// 4. 获取自动化配置
	automation, err := s.getAutomationProperties(ctx, flow, task)
	if err != nil {
		return nil, s.handleTaskError(ctx, task.Id, "提取自动化信息失败", domain.FAILED, err)
	}

	// 5. 获取调度节点
	runnerResp, err := s.getScheduleRunner(ctx, automation, orderResp)
	if err != nil {
		return nil, s.handleTaskError(ctx, task.Id, "获取调度节点失败", domain.PENDING, err)
	}

	// 6. 获取代码模板
	codebookResp, err := s.codebookSvc.FindByUid(ctx, runnerResp.CodebookUid)
	if err != nil {
		return nil, s.handleTaskError(ctx, task.Id, "获取任务模版失败", domain.FAILED, err)
	}

	return &taskProcessContext{
		order:      orderResp,
		inst:       inst,
		flow:       flow,
		automation: automation,
		runner:     runnerResp,
		codebook:   codebookResp,
	}, nil
}

func (s *service) assembleRuntimeArgs(ctx context.Context, orderResp order.Order) (map[string]interface{}, error) {
	// 获取基础表单参数
	args, err := s.prepareUserArgs(ctx, orderResp)
	if err != nil {
		return nil, err
	}

	// 获取用户审批提交的增量表单数据
	formValue, err := s.orderSvc.FindTaskFormsByOrderID(ctx, orderResp.Id)
	if err != nil {
		return nil, err
	}

	// 覆盖合并
	for _, value := range formValue {
		args[value.Key] = value.Value
	}

	return args, nil
}

func (s *service) getScheduleRunner(ctx context.Context, automation easyflow.AutomationProperty,
	orderResp order.Order) (runner.Runner, error) {

	// 动态路由：自动根据表单提交字段动态匹配调度节点
	if automation.Tag == "auto" {
		return s.autoDiscoverRunner(ctx, automation, orderResp)
	}

	// 静态路由：根据明确指定的标签 (比如 default 等) 查找对应的 Worker
	return s.runnerSvc.FindByCodebookUidAndTag(ctx, automation.CodebookUid, automation.Tag)
}

func (s *service) autoDiscoverRunner(ctx context.Context, automation easyflow.AutomationProperty,
	orderResp order.Order) (runner.Runner, error) {
	// 1. 获取该模版相关的全部自动发现规则探针
	discoveries, total, err := s.discoverySvc.ListByTemplateId(ctx, 0, 100, orderResp.TemplateId)
	if err != nil {
		return runner.Runner{}, err
	}
	if total == 0 {
		return runner.Runner{}, fmt.Errorf("该模版尚未配置可用的节点自动发现策略规则")
	}

	// 2. 根据工单表单中用户实际填写的业务数据，过滤出符合匹配特征的调度网关 Runner IDs
	matchedRunnerIDs := slice.FilterMap(discoveries, func(idx int, d discovery.Discovery) (int64, bool) {
		val, ok := orderResp.Data[d.Field]
		return d.RunnerId, ok && val == d.Value
	})

	if len(matchedRunnerIDs) == 0 {
		return runner.Runner{}, fmt.Errorf("当前工单填写的业务参数，未能触及任何匹配的自动发现策略")
	}

	// 3. 批量拉取符合特征的这些远端 Runner 实体
	runners, err := s.runnerSvc.ListByIds(ctx, matchedRunnerIDs)
	if err != nil {
		return runner.Runner{}, err
	}

	// 4. 从中筛选具备承接目前这个任务剧本 (CodebookUid) 能力的最终调配网关
	matchedRunner, found := slice.Find(runners, func(src runner.Runner) bool {
		return src.CodebookUid == automation.CodebookUid
	})

	if !found {
		return runner.Runner{}, fmt.Errorf("未能在匹配到的工作节点群落中，找到能够承载当前剧本(UID: %s)的可用网关", automation.CodebookUid)
	}

	return matchedRunner, nil
}

func (s *service) prepareTaskUpdate(orderResp order.Order, task domain.Task, flow workflow.Workflow,
	codebookResp codebook.Codebook, runnerResp runner.Runner, args map[string]interface{},
	status domain.Status, timing domain.Timing, automation easyflow.AutomationProperty) domain.Task {

	workerName := ""
	topic := ""

	if runnerResp.IsModeWorker() {
		workerName = runnerResp.Worker.WorkerName
		topic = runnerResp.Worker.Topic
	} else {
		workerName = runnerResp.Execute.Handler
		topic = runnerResp.Execute.ServiceName
	}

	return domain.Task{
		// 可选字段
		Id:            task.Id,
		WorkerName:    workerName,
		Topic:         topic,
		ProcessInstId: task.ProcessInstId,
		WorkflowId:    flow.Id,
		CodebookUid:   codebookResp.Identifier,
		CodebookName:  codebookResp.Name,

		// 必填字段
		OrderId:  orderResp.Id,
		Code:     codebookResp.Code,
		Language: codebookResp.Language,
		Status:   status,
		Args:     args,
		Variables: slice.Map(runnerResp.Variables, func(idx int, src runner.Variables) domain.Variables {
			return domain.Variables{
				Key:    src.Key,
				Value:  src.Value,
				Secret: src.Secret,
			}
		}),

		// 定时任务
		IsTiming: automation.IsTiming,
		Timing:   timing,
	}
}

func (s *service) prepareUserArgs(ctx context.Context, orderResp order.Order) (map[string]interface{}, error) {
	args := orderResp.Data
	userInfo, err := s.userSvc.FindByUsername(ctx, orderResp.CreateBy)
	if err != nil {
		s.logger.Error("获取用户信息失败", elog.FieldErr(err))
		return args, nil
	}

	userInfo.Password = "[Mask]"
	userInfoJSON, _ := json.Marshal(userInfo)
	args["user_info"] = string(userInfoJSON)
	return args, nil
}

func (s *service) getAutomationProperties(ctx context.Context, flow workflow.Workflow, task domain.Task) (
	easyflow.AutomationProperty, error) {
	return s.workflowSvc.GetAutomationProperty(easyflow.Workflow{
		Id:    flow.Id,
		Name:  flow.Name,
		Owner: flow.Owner,
		FlowData: easyflow.LogicFlow{
			Edges: flow.FlowData.Edges,
			Nodes: flow.FlowData.Nodes,
		},
	}, task.CurrentNodeId)
}

func (s *service) handleTaskError(ctx context.Context, taskID int64, triggerPosition string, status domain.Status, err error) error {
	if updateErr := s.updateTaskStatus(ctx, taskID, triggerPosition, status, err.Error()); updateErr != nil {
		s.logger.Error("更新任务状态失败", elog.FieldErr(updateErr))
	}
	return err
}

func (s *service) updateTaskStatus(ctx context.Context, taskID int64, triggerPosition string, status domain.Status, result string) error {
	_, err := s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
		Id:              taskID,
		TriggerPosition: triggerPosition,
		Status:          status,
		Result:          result,
	})
	return err
}

func (s *service) determineTaskStatus(automation easyflow.AutomationProperty, data map[string]interface{}) (domain.Status, domain.Timing) {
	status := domain.RUNNING
	timing := domain.Timing{Stime: time.Now().UnixMilli()}

	if automation.IsTiming {
		status = domain.TIMING
		timing = s.calculateTiming(automation, data)
	}
	return status, timing
}

func (s *service) calculateTiming(automation easyflow.AutomationProperty, data map[string]interface{}) domain.Timing {
	if !automation.IsTiming {
		return domain.Timing{}
	}

	timing := domain.Timing{
		Unit:     domain.HOUR,
		Quantity: 2,
	}

	// 1. 根据执行方式初始化调度单位和时长
	switch automation.ExecMethod {
	case "template":
		timing.Quantity = s.parseTemplateQuantity(automation.TemplateField, data)
	case "hand":
		timing.Unit = domain.Unit(automation.Unit)
		timing.Quantity = automation.Quantity
	}

	// 2. 根据单位计算时差并合成最终执行毫米级时间戳
	duration := s.calculateDuration(timing.Unit, timing.Quantity)
	timing.Stime = time.Now().Add(duration).UnixMilli()

	return timing
}

// parseTemplateQuantity 尝试从动态表单上下文中提取并转换为合法的时长 (int64)
func (s *service) parseTemplateQuantity(field string, data map[string]interface{}) int64 {
	const defaultQuantity = 2

	quantityVal, exist := data[field]
	if !exist {
		s.logger.Warn("字段不存在, 赋值默认值 2 H", elog.String("field", field))
		return defaultQuantity
	}

	switch v := quantityVal.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		parsedQuantity, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			s.logger.Error("解析失败, 赋值默认值 2 H", elog.FieldErr(err), elog.Any("value", v))
			return defaultQuantity
		}
		return parsedQuantity
	default:
		s.logger.Warn("类型未知, 赋值默认值 2 H", elog.Any("type", fmt.Sprintf("%T", v)), elog.Any("value", v))
		return defaultQuantity
	}
}

// calculateDuration 将业务域的时间单位转化为 Go 中的标准 time.Duration
func (s *service) calculateDuration(unit domain.Unit, quantity int64) time.Duration {
	switch unit {
	case domain.MINUTE:
		return time.Duration(quantity) * time.Minute
	case domain.DAY:
		return time.Duration(quantity) * 24 * time.Hour
	case domain.HOUR:
		return time.Duration(quantity) * time.Hour
	default:
		s.logger.Warn("未知的时间单位，按照小时进行计算", elog.Any("unit", unit))
		return time.Duration(quantity) * time.Hour
	}
}
