package easyflow

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order"
	templateSvc "github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/enotify/notify"
	"github.com/ecodeclub/ekit/retry"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"sync"
	"time"
)

type NotificationService interface {
	Send(ctx context.Context, instanceId int, userIDs []string) (bool, error)
}

type Notify struct {
	integrations []NotifyIntegration
	engineSvc    engineSvc.Service
	userSvc      user.Service
	templateSvc  templateSvc.Service
	orderSvc     order.Service
	workflowSvc  workflow.Service

	initialInterval time.Duration
	maxInterval     time.Duration
	maxRetries      int32
	logger          *elog.Component
}

func NewNotify(engineSvc engineSvc.Service, templateSvc templateSvc.Service, orderSvc order.Service,
	userSvc user.Service, workflowSvc workflow.Service, integrations []NotifyIntegration) (*Notify, error) {

	return &Notify{
		engineSvc:       engineSvc,
		templateSvc:     templateSvc,
		orderSvc:        orderSvc,
		userSvc:         userSvc,
		workflowSvc:     workflowSvc,
		logger:          elog.DefaultLogger,
		integrations:    integrations,
		initialInterval: 5 * time.Second,
		maxRetries:      int32(3),
		maxInterval:     15 * time.Second,
	}, nil
}

func (n *Notify) Send(ctx context.Context, instanceId int, userIDs []string) (bool, error) {
	// 查看是否需要发送消息， method 消息通知渠道
	o, method, er := n.isNotify(ctx, instanceId)
	if er != nil {
		n.logger.Error("跳过消息通知",
			elog.Any("error", er),
			elog.Any("instId", instanceId),
		)
		return false, nil
	}

	rules, er := n.getRules(ctx, o)
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
					elog.Any("instId", instanceId),
				)
				break
			}

			// 获取当前任务流转到的用户
			tasks, err = n.engineSvc.GetTasksByInstUsers(ctx, instanceId, userIDs)
			if err != nil || len(tasks) == 0 {
				time.Sleep(d)
				continue
			}

			break
		}

		// 获取用户的详情信息
		users, err := n.getUsers(ctx, tasks)
		if err != nil {
			n.logger.Error("用户查询失败",
				elog.FieldErr(err),
			)
		}

		// 生成消息数据
		var messages []notify.NotifierWrap
		for _, integration := range n.integrations {
			if integration.name == workflow.NotifyMethodToString(method) {
				messages = integration.notifier.builder(rules, o, users, tasks)
				break
			}
		}

		// 异步发送消息
		if ok, err := n.send(ctx, messages); err != nil || !ok {
			n.logger.Warn("发送消息失败",
				elog.Any("error", err),
				elog.Any("user_ids", userIDs),
			)
		}
	}()

	return true, nil
}

// getUsers 获取需要通知的用户信息
func (n *Notify) getUsers(ctx context.Context, tasks []model.Task) ([]user.User, error) {
	userIds := slice.Map(tasks, func(idx int, src model.Task) string {
		return src.UserID
	})

	users, err := n.userSvc.FindByUsernames(ctx, userIds)
	if err != nil {
		return nil, err
	}

	return users, err
}

// send 发送消息通知
func (n *Notify) send(ctx context.Context, notifyWrap []notify.NotifierWrap) (bool, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstError error
	success := true

	// 使用 goroutines 发送消息
	for _, msg := range notifyWrap {
		wg.Add(1)
		nw := msg
		go func(m notify.NotifierWrap) {
			defer wg.Done()

			ok, err := nw.Send(ctx)
			if err != nil {
				mu.Lock() // 锁定访问共享资源
				if firstError == nil {
					firstError = err // 记录第一个出现的错误
				}
				success = false
				mu.Unlock()
			}

			if !ok {
				mu.Lock()
				if firstError == nil {
					firstError = fmt.Errorf("消息发送失败")
				}
				success = false
				mu.Unlock()
			}
		}(msg)
	}

	// 等待所有 goroutine 完成
	wg.Wait()

	if firstError != nil {
		return false, firstError
	}
	return success, nil
}

// isNotify 判断是否开启消息通知
func (n *Notify) isNotify(ctx context.Context, InstanceId int) (order.Order, workflow.NotifyMethod, error) {
	// 获取工单详情信息
	o, err := n.orderSvc.DetailByProcessInstId(ctx, InstanceId)
	if err != nil {
		return o, 0, err
	}

	// 判断是否需要消息提示
	wf, err := n.workflowSvc.Find(ctx, o.WorkflowId)
	if err != nil {
		return o, 0, err
	}

	if !wf.IsNotify {
		return order.Order{}, 0, fmt.Errorf("流程控制未开启消息通知能力")
	}

	return o, wf.NotifyMethod, nil
}

// isNotify 获取模版的字段信息
func (n *Notify) getRules(ctx context.Context, order order.Order) ([]Rule, error) {
	// 获取模版详情信息
	t, err := n.templateSvc.DetailTemplate(ctx, order.TemplateId)
	if err != nil {
		return nil, err
	}

	rules, err := parseRules(t.Rules)
	if err != nil {
		return nil, err
	}

	return rules, nil
}

// parseRules 解析模版字段
func parseRules(ruleData interface{}) ([]Rule, error) {
	var rules []Rule
	rulesJson, err := json.Marshal(ruleData)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(rulesJson, &rules)
	return rules, err
}
