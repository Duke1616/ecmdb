package node

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/event/easyflow/notification"
	"github.com/Duke1616/ecmdb/internal/event/easyflow/notification/method"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	"github.com/Duke1616/ecmdb/internal/task"
	templateSvc "github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/enotify/notify"
	"github.com/ecodeclub/ekit/retry"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"time"
)

type UserNotification[T any] struct {
	integrations []method.NotifyIntegration
	engineSvc    engineSvc.Service
	taskSvc      task.Service
	userSvc      user.Service
	templateSvc  templateSvc.Service
	orderSvc     order.Service

	initialInterval time.Duration
	maxInterval     time.Duration
	maxRetries      int32
	logger          *elog.Component
}

func (n *UserNotification[T]) UnmarshalProperty(ctx context.Context, wf workflow.Workflow, nodeId string) (T, error) {
	var t T

	return t, nil
}

func NewUserNotification[T any](engineSvc engineSvc.Service, templateSvc templateSvc.Service, orderSvc order.Service,
	userSvc user.Service, taskSvc task.Service, integrations []method.NotifyIntegration) (*UserNotification[T], error) {

	return &UserNotification[T]{
		engineSvc:       engineSvc,
		templateSvc:     templateSvc,
		orderSvc:        orderSvc,
		userSvc:         userSvc,
		taskSvc:         taskSvc,
		logger:          elog.DefaultLogger,
		integrations:    integrations,
		initialInterval: 5 * time.Second,
		maxRetries:      int32(3),
		maxInterval:     15 * time.Second,
	}, nil
}

func (n *UserNotification[T]) Send(ctx context.Context, nOrder order.Order, params notification.NotifyParams) (bool, error) {
	rules, er := n.getRules(ctx, nOrder)
	if er != nil {
		return false, er
	}

	variables, er := engine.ResolveVariables(params.InstanceId, []string{"$starter"})
	if er != nil {
		return false, er
	}

	startUser, er := n.userSvc.FindByUsername(ctx, variables["$starter"])
	if er != nil {
		return false, er
	}

	// 只有当 Event 结束才能正确获取到 TaskId 信息，放到 Go Routine 异步运行, 通过重试机制获取到数据
	go func() {
		strategy, err := retry.NewExponentialBackoffRetryStrategy(n.initialInterval, n.maxInterval, n.maxRetries)
		if err != nil {
			return
		}

		var tasks []model.Task
		for {
			d, ok := strategy.Next()
			if !ok {
				n.logger.Error("处理执行任务超过最大重试次数",
					elog.Any("error", err),
					elog.Any("instId", params.InstanceId),
				)
				break
			}

			// 获取当前任务流转到的用户
			tasks, err = n.engineSvc.GetTasksByInstUsers(context.Background(), params.InstanceId, params.UserIDs)
			if err != nil || len(tasks) == 0 {
				time.Sleep(d)
				continue
			}

			break
		}

		// 获取用户的详情信息
		users, err := n.getUsers(context.Background(), tasks)
		if err != nil {
			n.logger.Error("用户查询失败",
				elog.FieldErr(err),
			)
		}

		// 生成消息数据
		title := rule.GenerateTitle(startUser.DisplayName, nOrder.TemplateName)
		var messages []notify.NotifierWrap
		for _, integration := range n.integrations {
			if integration.Name == fmt.Sprintf("%s_%s", params.NotifyMethod, "user") {
				messages = integration.Notifier.Builder(title, users, method.NotifyParams{
					Order:      nOrder,
					WantResult: params.WantResult,
					Tasks:      tasks,
					Rules:      rules,
				})
				break
			}
		}

		// 异步发送消息
		var ok bool
		if ok, err = send(context.Background(), messages); err != nil || !ok {
			n.logger.Warn("发送消息失败",
				elog.Any("error", err),
				elog.Any("user_ids", params.UserIDs),
			)
		}
	}()

	return true, nil
}

func (n *UserNotification[T]) IsNotification(ctx context.Context, wf workflow.Workflow, instanceId int,
	nodeId string) (bool, map[string]interface{}, error) {
	if !wf.IsNotify {
		return false, nil, nil
	}

	nodesJSON, err := json.Marshal(wf.FlowData.Nodes)
	if err != nil {
		return false, nil, err
	}
	var nodes []easyflow.Node
	err = json.Unmarshal(nodesJSON, &nodes)
	if err != nil {
		return false, nil, err
	}

	mergedResult := make(map[string]interface{})
	for _, node := range nodes {
		switch node.Type {
		case "automation":
			property, _ := easyflow.ToNodeProperty[easyflow.AutomationProperty](node)
			if !property.IsNotify || property.NotifyMethod != ProcessEndSend {
				continue
			}

			result, er := n.taskSvc.FindTaskResult(ctx, instanceId, node.ID)
			if er != nil {
				return false, nil, er
			}

			if result.WantResult == "" {
				continue
			}

			var wantResult map[string]interface{}
			er = json.Unmarshal([]byte(result.WantResult), &wantResult)
			if er != nil {
				return false, nil, er
			}

			for key, value := range wantResult {
				mergedResult[key] = value
			}
			//case "user":
			//	property, _ := easyflow.ToNodeProperty[easyflow.UserProperty](node)
		}
	}

	if len(mergedResult) == 0 {
		return true, nil, nil
	}

	return true, mergedResult, nil
}

// getUsers 获取需要通知的用户信息
func (n *UserNotification[T]) getUsers(ctx context.Context, tasks []model.Task) ([]user.User, error) {
	userIds := slice.Map(tasks, func(idx int, src model.Task) string {
		return src.UserID
	})

	users, err := n.userSvc.FindByUsernames(ctx, userIds)
	if err != nil {
		return nil, err
	}

	return users, err
}

// isNotify 获取模版的字段信息
func (n *UserNotification[T]) getRules(ctx context.Context, order order.Order) ([]rule.Rule, error) {
	// 获取模版详情信息
	t, err := n.templateSvc.DetailTemplate(ctx, order.TemplateId)
	if err != nil {
		return nil, err
	}

	rules, err := rule.ParseRules(t.Rules)

	if err != nil {
		return nil, err
	}

	return rules, nil
}
