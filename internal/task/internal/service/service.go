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
	"github.com/Duke1616/ecmdb/internal/task/internal/service/dispatch"
	"github.com/Duke1616/ecmdb/internal/task/internal/service/scheduler"
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

	// RetryTask 人工手动触发重试，会重置重试计数器
	RetryTask(ctx context.Context, id int64) error

	// AutoRetryTask 背景定时任务自动补发，会累加重试计数器
	AutoRetryTask(ctx context.Context, id int64) error

	// UpdateTaskStatus 被底层异步执行通道回调更新回调结果
	UpdateTaskStatus(ctx context.Context, req domain.TaskResult) (int64, error)

	// UpdateArgs 动态修改参数上下文环境信息
	UpdateArgs(ctx context.Context, id int64, args map[string]interface{}) (int64, error)

	// UpdateVariables 修改执行环境变量内容（带有敏感字段防篡改规则）
	UpdateVariables(ctx context.Context, id int64, variables []domain.Variables) (int64, error)

	// ListTaskByStatus 用于平台展示过滤指定生命周期下的节点任务
	ListTaskByStatus(ctx context.Context, offset, limit int64, status uint8) ([]domain.Task, int64, error)

	// ListTaskByStatusAndMode 用于过滤指定生命周期和运行模式的任务
	ListTaskByStatusAndMode(ctx context.Context, offset, limit int64, status uint8, mode string) ([]domain.Task, int64, error)

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

	// UpdateExternalId 绑定外部分布式平台的任务 ID
	UpdateExternalId(ctx context.Context, id int64, externalId string) error
}

type service struct {
	repo         repository.TaskRepository
	scheduler    scheduler.Scheduler
	logger       *elog.Component
	orderSvc     order.Service
	userSvc      user.Service
	discoverySvc discovery.Service
	engineSvc    engine.Service
	workflowSvc  workflow.Service
	codebookSvc  codebook.Service
	runnerSvc    runner.Service
	dispatcher   dispatch.TaskDispatcher
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

func (s *service) ListTaskByStatusAndMode(ctx context.Context, offset, limit int64, status uint8, mode string) ([]domain.Task, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Task
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListTaskByStatusAndMode(ctx, offset, limit, status, mode)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalByStatusAndMode(ctx, status, mode)
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

func (s *service) UpdateExternalId(ctx context.Context, id int64, externalId string) error {
	return s.repo.UpdateExternalId(ctx, id, externalId)
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
	codebookSvc codebook.Service, runnerSvc runner.Service, engineSvc engine.Service,
	userSvc user.Service, dispatcher dispatch.TaskDispatcher, discoverySvc discovery.Service,
	sch scheduler.Scheduler) Service {
	return &service{
		repo:         repo,
		logger:       elog.DefaultLogger,
		orderSvc:     orderSvc,
		workflowSvc:  workflowSvc,
		codebookSvc:  codebookSvc,
		runnerSvc:    runnerSvc,
		engineSvc:    engineSvc,
		userSvc:      userSvc,
		dispatcher:   dispatcher,
		discoverySvc: discoverySvc,
		scheduler:    sch,
	}
}

func (s *service) StartTask(ctx context.Context, processInstId int, nodeId string) error {
	// 避免并发双写和流程空洞：原子化创建或查找当前流程的任务挂载点
	task, err := s.repo.FindOrCreate(ctx, domain.Task{
		ProcessInstId:   processInstId,
		CurrentNodeId:   nodeId,
		TriggerPosition: "准备启动节点",
		Status:          domain.WAITING,
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
	// NOTE: 人工点击重试，首先重置重试计数器为 0
	if _, err := s.UpdateTaskStatus(ctx, domain.TaskResult{
		Id:              id,
		TriggerPosition: "人工触发手动重试",
		RetryCount:      -1, // 特殊约定：-1 表示重置为 0
	}); err != nil {
		s.logger.Error("手动重试重置计数器失败", elog.Int64("taskId", id), elog.FieldErr(err))
	}

	task, err := s.repo.FindById(ctx, id)
	if err != nil {
		elog.Error("获取任务失败", elog.FieldErr(err), elog.Int64("taskId", id))
		return err
	}

	// 证明这个任务是失败的，应该重新获取数据
	if task.OrderId == 0 {
		return s.process(ctx, task)
	}

	return s.retry(ctx, task, false)
}

func (s *service) AutoRetryTask(ctx context.Context, id int64) error {
	task, err := s.repo.FindById(ctx, id)
	if err != nil {
		return err
	}

	// 自动重试需要增加计数，并受 MaxRetry 限制
	return s.retry(ctx, task, true)
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
	// 如果状态发生流转变迁且不再是待调度类别的，清理它被加上的调度内存锁，这样以后一旦 Retry 断线重发等机制激活，才能再次派发成功。
	if req.Status != domain.SCHEDULED && req.Status != domain.WAITING {
		s.scheduler.Remove(req.Id)
	}

	// NOTE: 自动补充时间戳，避免上层调用方分散管理时间记录
	now := time.Now().UnixMilli()
	switch req.Status {
	case domain.RUNNING:
		// 任务正式开始执行，记录真实开始时间
		req.StartTime = now
	case domain.SUCCESS, domain.FAILED:
		// 任务进入终态，记录结束时间
		req.EndTime = now
	}

	return s.repo.UpdateTaskStatus(ctx, req)
}

func (s *service) retry(ctx context.Context, task domain.Task, auto bool) error {
	if auto {
		const maxRetryCount = 5
		if task.RetryCount >= maxRetryCount {
			s.logger.Warn("任务自动重试次数已达上限，转为 BLOCKED 等待人工介入",
				elog.Int64("taskId", task.Id),
				elog.Int("retryCount", task.RetryCount),
			)
			_, _ = s.UpdateTaskStatus(ctx, domain.TaskResult{
				Id:              task.Id,
				TriggerPosition: "自动恢复失败：超过最大重试次数",
				Status:          domain.BLOCKED,
			})
			return nil
		}
	}

	res := domain.TaskResult{
		Id:              task.Id,
		TriggerPosition: "自动补发任务",
		Status:          domain.SCHEDULED,
	}
	if !auto {
		res.TriggerPosition = "人工手动重试"
	} else {
		res.RetryCount = 1
	}

	_, _ = s.UpdateTaskStatus(ctx, res)

	return s.dispatchTask(ctx, task)
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
	return s.dispatchTask(ctx, taskUpdate)
}

func (s *service) dispatchTask(ctx context.Context, task domain.Task) error {
	// 对所有任务都使用内存锁防重，而不只是定时任务。
	// 原因：非定时任务在 Dispatch 成功到 UpdateTaskStatus(RUNNING) 写库完成之间存在时间窗口，
	// 若此期间 offine_recovery 轮询，会误判状态仍为 SCHEDULED 并触发重复派发。
	// 锁会在 UpdateTaskStatus 将状态变更为非调度态时（RUNNING/SUCCESS/FAILED）自动 Remove。
	if !s.scheduler.Add(task.Id) {
		s.logger.Info("任务正在派发中或已在内存调度组，跳过重复派发", elog.Int64("taskId", task.Id))
		return nil
	}

	if err := s.dispatcher.Dispatch(ctx, task); err != nil {
		// 派发失败，释放锁，以便下一次重试轮询可以继续调度
		s.scheduler.Remove(task.Id)
		return err
	}

	// === 统一在此处变更状态 === //
	// 只要 Dispatch 没报错，意味着无论是入库队列还是发远端，任务都已正式“上路”。
	// 只有当任务是 Worker 模式且为本地定时任务时，才保留在 SCHEDULED 状态让本地内存 Cron 调度及宕机恢复。
	if !task.IsTiming || task.RunMode == domain.RunModeExecute {
		// 此处若更新失败，任务实际已路由到执行端，状态库仍停在 SCHEDULED，
		if _, statusErr := s.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			Status:          domain.RUNNING,
			TriggerPosition: "分发已送达执行端，当前任务执行中",
		}); statusErr != nil {
			s.logger.Error("任务已派发但状态更新 RUNNING 失败，存在重复调度风险",
				elog.Int64("taskId", task.Id),
				elog.FieldErr(statusErr),
			)
		}
	}

	return nil
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
		return nil, s.handleTaskError(ctx, task.Id, "获取调度节点失败", domain.BLOCKED, err)
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

	t := domain.Task{
		Id:              task.Id,
		ProcessInstId:   task.ProcessInstId,
		OrderId:         orderResp.Id,
		WorkflowId:      flow.Id,
		CodebookUid:     codebookResp.Identifier,
		CodebookName:    codebookResp.Name,
		Code:            codebookResp.Code,
		Language:        codebookResp.Language,
		Status:          status,
		Args:            args,
		TriggerPosition: "准备启动节点",
		IsTiming:        automation.IsTiming,
		Timing:          timing,
		Variables:       s.toDomainVariables(runnerResp.Variables),
	}

	// 填充模式差异化运行参数
	s.applyRunnerConfig(&t, runnerResp)

	return t
}

func (s *service) applyRunnerConfig(t *domain.Task, r runner.Runner) {
	if r.IsModeWorker() {
		t.RunMode = domain.RunModeWorker
		t.WorkerName = r.Worker.WorkerName
		t.Topic = r.Worker.Topic
		t.Worker = &domain.Worker{
			WorkerName: r.Worker.WorkerName,
			Topic:      r.Worker.Topic,
		}
		return
	}

	// 默认为分布式执行模式 (EXECUTE)
	t.RunMode = domain.RunModeExecute
	t.WorkerName = r.Execute.Handler
	t.Topic = r.Execute.ServiceName
	t.Execute = &domain.Execute{
		ServiceName: r.Execute.ServiceName,
		Handler:     r.Execute.Handler,
	}
}

func (s *service) toDomainVariables(vars []runner.Variables) []domain.Variables {
	return slice.Map(vars, func(idx int, src runner.Variables) domain.Variables {
		return domain.Variables{
			Key:    src.Key,
			Value:  src.Value,
			Secret: src.Secret,
		}
	})
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
	_, updateErr := s.UpdateTaskStatus(ctx, domain.TaskResult{
		Id:              taskID,
		TriggerPosition: triggerPosition,
		Status:          status,
		Result:          err.Error(),
	})
	if updateErr != nil {
		s.logger.Error("更新任务状态失败", elog.FieldErr(updateErr))
	}
	return err
}

func (s *service) determineTaskStatus(automation easyflow.AutomationProperty, data map[string]interface{}) (domain.Status, domain.Timing) {
	status := domain.SCHEDULED
	timing := domain.Timing{Stime: time.Now().UnixMilli()}

	if automation.IsTiming {
		status = domain.SCHEDULED
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
