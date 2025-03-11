package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/runner"
	"github.com/Duke1616/ecmdb/internal/task/internal/domain"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"github.com/spf13/viper"
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

	ListSuccessTasksByCtime(ctx context.Context, offset, limit int64, ctime int64) ([]domain.Task, int64, error)

	// FindTaskResult 查找自动化任务
	FindTaskResult(ctx context.Context, instanceId int, nodeId string) (domain.Task, error)

	// Detail 查看任务信息
	Detail(ctx context.Context, id int64) (domain.Task, error)
}

type service struct {
	repo        repository.TaskRepository
	aesKey      string
	logger      *elog.Component
	orderSvc    order.Service
	userSvc     user.Service
	engineSvc   engine.Service
	workflowSvc workflow.Service
	codebookSvc codebook.Service
	runnerSvc   runner.Service
	workerSvc   worker.Service
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
	codebookSvc codebook.Service, runnerSvc runner.Service, workerSvc worker.Service, engineSvc engine.Service,
	userSvc user.Service) Service {
	return &service{
		repo:        repo,
		aesKey:      viper.Get("crypto_aes_key").(string),
		logger:      elog.DefaultLogger,
		orderSvc:    orderSvc,
		workflowSvc: workflowSvc,
		codebookSvc: codebookSvc,
		runnerSvc:   runnerSvc,
		workerSvc:   workerSvc,
		engineSvc:   engineSvc,
		userSvc:     userSvc,
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

	// 变量数据梳理
	vars := slice.Map(task.Variables, func(idx int, src domain.Variables) domain.Variables {
		if src.Secret {
			val, er := cryptox.DecryptAES[any](s.aesKey, src.Value.(string))
			if er != nil {
				return domain.Variables{}
			}

			return domain.Variables{
				Key:    src.Key,
				Value:  val,
				Secret: src.Secret,
			}
		}

		return domain.Variables{
			Key:    src.Key,
			Value:  src.Value,
			Secret: src.Secret,
		}
	})

	// 添加工单创建人
	variables, _ := json.Marshal(vars)
	return s.workerSvc.Execute(ctx, worker.Execute{
		TaskId:    task.Id,
		Topic:     task.Topic,
		Code:      task.Code,
		Language:  task.Language,
		Args:      task.Args,
		Variables: string(variables),
	})
}

func (s *service) process(ctx context.Context, task domain.Task) error {
	// 获取工单信息
	orderResp, err := s.orderSvc.DetailByProcessInstId(ctx, task.ProcessInstId)
	if err != nil {
		_, _ = s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			TriggerPosition: "获取工单失败",
			Status:          domain.FAILED,
			Result:          err.Error(),
		})
		return err
	}

	// TODO 后期引用[实时] OR [定时]执行逻辑  目前都是立即执行
	flow, err := s.workflowSvc.Find(ctx, orderResp.WorkflowId)
	if err != nil {
		_, _ = s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			TriggerPosition: "获取流程信息失败",
			Status:          domain.FAILED,
			Result:          err.Error(),
		})
		return err
	}

	// 获取自动化提交信息
	automation, err := s.workflowSvc.GetAutomationProperty(easyflow.Workflow{
		Id:    flow.Id,
		Name:  flow.Name,
		Owner: flow.Owner,
		FlowData: easyflow.LogicFlow{
			Edges: flow.FlowData.Edges,
			Nodes: flow.FlowData.Nodes,
		},
	}, task.CurrentNodeId)

	if err != nil {
		_, _ = s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			TriggerPosition: "提取自动化信息失败",
			Status:          domain.FAILED,
			Result:          err.Error(),
		})
		return err
	}

	// 查看调度工作节点
	runnerResp, err := s.runnerSvc.FindByCodebookUid(ctx, automation.CodebookUid, automation.Tag)
	if err != nil {
		_, _ = s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			TriggerPosition: "获取调度节点失败",
			Status:          domain.PENDING,
			Result:          err.Error(),
		})
		return err
	}

	// 查看工作节点Topic
	workerResp, err := s.workerSvc.FindByName(ctx, runnerResp.WorkerName)
	if err != nil {
		_, _ = s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			TriggerPosition: "获取工作节点失败",
			Status:          domain.FAILED,
			Result:          err.Error(),
		})
		return err
	}

	// 查询执行代码
	codebookResp, err := s.codebookSvc.FindByUid(ctx, runnerResp.CodebookUid)
	if err != nil {
		_, _ = s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			TriggerPosition: "获取任务模版失败",
			Status:          domain.FAILED,
			Result:          err.Error(),
		})
		return err
	}

	// 添加工单创建人
	args := orderResp.Data
	userInfo, err := s.userSvc.FindByUsername(ctx, orderResp.CreateBy)
	if err != nil {
		s.logger.Error("获取用户信息失败，可能系统中不存在", elog.FieldErr(err))
	}
	userInfo.Password = "[Mask]"
	userInfoJSON, err := json.Marshal(userInfo)
	args["user_info"] = string(userInfoJSON)

	// 根据条件设置状态
	status := domain.RUNNING
	timing := domain.Timing{Stime: time.Now().UnixMilli()}
	if automation.IsTiming {
		status = domain.TIMING
		switch automation.ExecMethod {
		case "template":
			timing.Unit = domain.HOUR
			quantity, exist := orderResp.Data[automation.TemplateField]
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

	// 初始化任务更新数据
	taskUpdate := domain.Task{
		// 可选字段
		Id:            task.Id,
		ProcessInstId: task.ProcessInstId,
		WorkerName:    workerResp.Name,
		WorkflowId:    flow.Id,
		CodebookUid:   codebookResp.Identifier,
		CodebookName:  codebookResp.Name,

		// 必填字段
		OrderId:  orderResp.Id,
		Code:     codebookResp.Code,
		Topic:    workerResp.Topic,
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

	// 更新任务
	_, err = s.repo.UpdateTask(ctx, taskUpdate)
	if err != nil {
		return err
	}

	// 如果是定时任务，直接返回
	if automation.IsTiming {
		s.logger.Info("定时执行， 退出当前步骤")
		return nil
	}

	// TODO 查看节点状态，禁用 离线 是否可以堆积到消息队列中
	switch workerResp.Status {
	case worker.STOPPING:
	case worker.OFFLINE:
		_, _ = s.repo.UpdateTaskStatus(ctx, domain.TaskResult{
			Id:              task.Id,
			TriggerPosition: "调度任务节点失败, 工作节点离线",
			Status:          domain.FAILED,
			Result:          "调度任务节点失败, 工作节点离线",
		})
		return err
	}

	// 变量数据梳理
	variables := slice.Map(runnerResp.Variables, func(idx int, src runner.Variables) domain.Variables {
		if src.Secret {
			val, er := cryptox.DecryptAES[any](s.aesKey, src.Value.(string))
			if er != nil {
				return domain.Variables{}
			}

			return domain.Variables{
				Key:    src.Key,
				Value:  val,
				Secret: src.Secret,
			}
		}

		return domain.Variables{
			Key:    src.Key,
			Value:  src.Value,
			Secret: src.Secret,
		}
	})

	// 运行任务
	vars, _ := json.Marshal(variables)
	return s.workerSvc.Execute(ctx, worker.Execute{
		TaskId:    task.Id,
		Topic:     workerResp.Topic,
		Code:      codebookResp.Code,
		Language:  codebookResp.Language,
		Args:      args,
		Variables: string(vars),
	})
}
