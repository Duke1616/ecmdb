package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
	"strconv"
	"time"
)

type Service interface {
	// CreateTask 创建任务
	CreateTask(ctx context.Context, processInstId int, nodeId string) error
	// StartTask 启动任务
	StartTask(ctx context.Context, processInstId int, nodeId string) error
	RetryTask(ctx context.Context, id int64) error
	UpdateTaskStatus(ctx context.Context, req domain.TaskResult) (int64, error)
	UpdateArgs(ctx context.Context, id int64, args map[string]interface{}) (int64, error)
	UpdateVariables(ctx context.Context, id int64, variables []domain.Variables) (int64, error)
	// ListTaskByStatus 列表任务
	ListTaskByStatus(ctx context.Context, offset, limit int64, status uint8) ([]domain.Task, int64, error)
	ListTask(ctx context.Context, offset, limit int64) ([]domain.Task, int64, error)
	ListTaskByInstanceId(ctx context.Context, offset, limit int64, instanceId int) ([]domain.Task, int64, error)

	ListSuccessTasksByCtime(ctx context.Context, offset, limit int64, ctime int64) ([]domain.Task, int64, error)

	// FindTaskResult 查找自动化任务
	FindTaskResult(ctx context.Context, instanceId int, nodeId string) (domain.Task, error)

	// Detail 查看任务信息
	Detail(ctx context.Context, id int64) (domain.Task, error)
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

func (s *service) ListSuccessTasksByCtime(ctx context.Context, offset, limit int64, ctime int64) ([]domain.Task, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Task
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListSuccessTasksByCtime(ctx, offset, limit, ctime)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalByCtime(ctx, ctime)
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

func (s *service) CreateTask(ctx context.Context, processInstId int, nodeId string) error {
	_, err := s.repo.CreateTask(ctx, domain.Task{
		ProcessInstId:   processInstId,
		TriggerPosition: "任务等待",
		CurrentNodeId:   nodeId,
		Status:          domain.WAITING,
	})

	return err
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
	// 先创建任务、以防后续失败，导致无法溯源
	task, err := s.repo.FindByProcessInstId(ctx, processInstId, nodeId)
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return s.process(ctx, task)
	}

	// 新建工作记录
	taskId, err := s.repo.CreateTask(ctx, domain.Task{
		ProcessInstId:   processInstId,
		TriggerPosition: "开始节点",
		CurrentNodeId:   nodeId,
		Status:          domain.SCHEDULE,
	})

	// 追加赋值
	task.Id = taskId
	task.ProcessInstId = processInstId
	task.CurrentNodeId = nodeId

	if err != nil {
		elog.Error("创建任务失败",
			elog.Any("错误信息", err),
			elog.Any("流程实例ID", processInstId),
			elog.Any("当前节点ID", nodeId),
		)
		return err
	}

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

func (s *service) process(ctx context.Context, task domain.Task) error {
	// 1. 获取工单信息
	orderResp, err := s.orderSvc.DetailByProcessInstId(ctx, task.ProcessInstId)
	if err != nil {
		return s.handleTaskError(ctx, task.Id, "获取工单失败", domain.FAILED, err)
	}

	// 2. 获取流程信息
	flow, err := s.workflowSvc.Find(ctx, orderResp.WorkflowId)
	if err != nil {
		return s.handleTaskError(ctx, task.Id, "获取流程信息失败", domain.FAILED, err)
	}

	// 3. 获取自动化配置
	automation, err := s.getAutomationProperties(ctx, flow, task)
	if err != nil {
		return s.handleTaskError(ctx, task.Id, "提取自动化信息失败", domain.FAILED, err)
	}

	// 4. 获取调度节点
	runnerResp, err := s.getScheduleRunner(ctx, automation, orderResp)
	if err != nil {
		return s.handleTaskError(ctx, task.Id, "获取调度节点失败", domain.PENDING, err)
	}

	// 5. 获取代码模板
	codebookResp, err := s.codebookSvc.FindByUid(ctx, runnerResp.CodebookUid)
	if err != nil {
		return s.handleTaskError(ctx, task.Id, "获取任务模版失败", domain.FAILED, err)
	}

	// 6. 准备用户参数
	args, err := s.prepareUserArgs(ctx, orderResp)
	if err != nil {
		return err
	}

	// 7. 确定任务状态和定时配置
	status, timing := s.determineTaskStatus(automation, orderResp.Data)

	// 8. 构建任务更新数据
	taskUpdate := s.prepareTaskUpdate(orderResp, task, flow, codebookResp, runnerResp, args,
		status, timing, automation)

	// 9. 更新任务
	_, err = s.repo.UpdateTask(ctx, taskUpdate)
	if err != nil {
		return err
	}

	// 10. 处理定时任务
	if automation.IsTiming {
		go func() {
			_ = s.cronjobSvc.Add(ctx, taskUpdate)
		}()
		return nil
	}

	// TODO 查看节点状态，禁用 离线 是否可以堆积到消息队列中
	// TODO 暂时不考虑离线情况
	//switch workerResp.Status {
	//case worker.STOPPING:
	//case worker.OFFLINE:
	//	_, _ = s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
	//		Id:              task.Id,
	//		TriggerPosition: "调度任务节点失败, 工作节点离线",
	//		Status:          domain.FAILED,
	//		Result:          "调度任务节点失败, 工作节点离线",
	//	})
	//	return err
	//}

	return s.execSvc.Execute(ctx, task)
}

func (s *service) getScheduleRunner(ctx context.Context, automation easyflow.AutomationProperty,
	orderResp order.Order) (runner.Runner, error) {
	// 如果配置自动发现、将自动寻找匹配的调度节点
	if automation.Tag == "auto" {
		ds, total, err := s.discoverySvc.ListByTemplateId(ctx, 0, 100, orderResp.TemplateId)
		if err != nil {
			return runner.Runner{}, err
		}

		if total == 0 {
			return runner.Runner{}, fmt.Errorf("没有自动发现的规则")
		}

		ids := make([]int64, 0)
		for _, d := range ds {
			val, ok := orderResp.Data[d.Field]
			if !ok {
				continue
			}

			if val == d.Value {
				ids = append(ids, d.RunnerId)
			}
		}

		if len(ids) == 0 {
			return runner.Runner{}, fmt.Errorf("没有匹配自动发现规则")
		}

		rs, err := s.runnerSvc.ListByIds(ctx, ids)
		if err != nil {
			return runner.Runner{}, err
		}

		for _, r := range rs {
			if r.CodebookUid == automation.CodebookUid {
				return r, nil
			}
		}
	}

	return s.runnerSvc.FindByCodebookUid(ctx, automation.CodebookUid, automation.Tag)
}

func (s *service) prepareTaskUpdate(orderResp order.Order, task domain.Task, flow workflow.Workflow,
	codebookResp codebook.Codebook, runnerResp runner.Runner, args map[string]interface{},
	status domain.Status, timing domain.Timing, automation easyflow.AutomationProperty) domain.Task {

	return domain.Task{
		// 可选字段
		Id:            task.Id,
		ProcessInstId: task.ProcessInstId,
		WorkerName:    runnerResp.WorkerName,
		WorkflowId:    flow.Id,
		CodebookUid:   codebookResp.Identifier,
		CodebookName:  codebookResp.Name,

		// 必填字段
		OrderId:  orderResp.Id,
		Code:     codebookResp.Code,
		Topic:    runnerResp.Topic,
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
	timing := domain.Timing{}
	if automation.IsTiming {
		switch automation.ExecMethod {
		case "template":
			timing.Unit = domain.HOUR
			quantity, exist := data[automation.TemplateField]
			if !exist {
				s.logger.Warn("字段不存在, 赋值默认值 2 H")
				timing.Quantity = 2
				break
			}

			switch v := quantity.(type) {
			case int64:
				timing.Quantity = v
			case int:
				timing.Quantity = int64(v)
			case float64:
				timing.Quantity = int64(v)
			case string:
				parsedQuantity, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					s.logger.Error("解析失败, 赋值默认值 2 H",
						elog.FieldErr(err),
						elog.Any("value", v),
					)
					timing.Quantity = 2
				} else {
					timing.Quantity = parsedQuantity
				}
			default:
				s.logger.Warn("类型未知, 赋值默认值 2 H",
					elog.Any("type", v),
					elog.Any("value", quantity),
				)
				timing.Quantity = 2
			}
		case "hand":
			timing.Unit = domain.Unit(automation.Unit)
			timing.Quantity = automation.Quantity
		}
	}

	// 解析 stime（Unix 时间戳，单位：毫秒）
	stime := time.UnixMilli(time.Now().UnixMilli())

	// 计算时间差
	var duration time.Duration
	switch timing.Unit {
	case domain.MINUTE: // 分钟
		duration = time.Duration(timing.Quantity) * time.Minute
	case domain.HOUR: // 小时
		duration = time.Duration(timing.Quantity) * time.Hour
	case domain.DAY: // 天
		duration = time.Duration(timing.Quantity) * 24 * time.Hour
	default:
		duration = time.Duration(timing.Quantity) * time.Hour
		s.logger.Warn("未知的时间单位，按照小时进行计算")
	}

	// 计算开始执行的时间
	timing.Stime = stime.Add(duration).UnixMilli()

	return timing
}
