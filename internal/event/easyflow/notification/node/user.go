package node

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
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

type UserNotification struct {
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

func NewUserNotification(engineSvc engineSvc.Service, templateSvc templateSvc.Service, orderSvc order.Service,
	userSvc user.Service, taskSvc task.Service, integrations []method.NotifyIntegration) (*UserNotification, error) {

	return &UserNotification{
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

func (n *UserNotification) isNotify(wf workflow.Workflow, instanceId int) bool {
	if !wf.IsNotify {
		n.logger.Warn("流程控制未开启消息通知能力",
			elog.Any("instId", instanceId),
		)
		return false
	}

	return true
}

func (n *UserNotification) Unmarshal(wf workflow.Workflow) ([]easyflow.Node, error) {
	nodesJSON, err := json.Marshal(wf.FlowData.Nodes)
	if err != nil {
		return nil, err
	}
	var nodes []easyflow.Node
	err = json.Unmarshal(nodesJSON, &nodes)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (n *UserNotification) Send(ctx context.Context, nOrder order.Order, wf workflow.Workflow,
	instanceId int, nodeId string) (bool, error) {
	// 判断是否开启
	if ok := n.isNotify(wf, instanceId); !ok {
		return false, nil
	}

	// 获取流程节点 nodes 信息
	nodes, err := n.Unmarshal(wf)
	if err != nil {
		return false, err
	}

	// 获取当前节点信息
	property, err := n.getProperty(nodes, nodeId)
	if err != nil {
		return false, err
	}

	// 获取自动化任务执行结果
	wantResult, err := n.wantAllResult(ctx, instanceId, nodes)
	if err != nil {
		return false, err
	}

	// 解析配置
	rules, err := n.getRules(ctx, nOrder)
	if err != nil {
		return false, err
	}

	variables, err := engine.ResolveVariables(instanceId, []string{"$starter"})
	if err != nil {
		return false, err
	}

	startUser, err := n.userSvc.FindByUsername(ctx, variables["$starter"])
	if err != nil {
		return false, err
	}

	// 只有当 Event 结束才能正确获取到 TaskId 信息，放到 Go Routine 异步运行, 通过重试机制获取到数据
	go func() {
		strategy, er := retry.NewExponentialBackoffRetryStrategy(n.initialInterval, n.maxInterval, n.maxRetries)
		if er != nil {
			return
		}

		var tasks []model.Task
		for {
			d, ok := strategy.Next()
			if !ok {
				n.logger.Error("处理执行任务超过最大重试次数",
					elog.Any("error", er),
					elog.Any("instId", instanceId),
				)
				break
			}

			// 获取当前任务流转到的用户
			tasks, err = n.engineSvc.GetTasksByCurrentNodeId(context.Background(), instanceId, nodeId)
			if err != nil || len(tasks) == 0 {
				time.Sleep(d)
				continue
			}

			break
		}

		// 获取用户的详情信息
		users, er := n.getUsers(context.Background(), tasks)
		if er != nil {
			n.logger.Error("用户查询失败",
				elog.FieldErr(er),
			)
		}

		// 生成消息数据
		var messages []notify.NotifierWrap
		title := rule.GenerateTitle(startUser.DisplayName, nOrder.TemplateName)
		template := method.FeishuTemplateApprovalName

		// 判断如果是抄送情况
		if property.IsCC {
			template = method.FeishuTemplateCC
			title = rule.GenerateCCTitle(startUser.DisplayName, nOrder.TemplateName)

			// 处理自动通过
			go n.ccPass(ctx, tasks)
		}
		for _, integration := range n.integrations {
			if integration.Name == fmt.Sprintf("%s_%s", workflow.NotifyMethodToString(wf.NotifyMethod), "user") {
				messages = integration.Notifier.Builder(title, users, template, method.NotifyParams{
					Order:      nOrder,
					WantResult: wantResult,
					Tasks:      tasks,
					Rules:      rules,
				})
				break
			}
		}

		// 异步发送消息
		var ok bool
		if ok, er = send(context.Background(), messages); er != nil || !ok {
			n.logger.Warn("发送消息失败",
				elog.Any("error", er),
			)
		}
	}()

	return true, nil
}

func (n *UserNotification) ccPass(ctx context.Context, tasks []model.Task) {
	for _, t := range tasks {
		// 如果是非会签节点，处理一次直接退出
		if t.IsCosigned != 1 {
			err := n.engineSvc.Pass(ctx, t.TaskID, "自处理抄送节点审批")
			if err != nil {
				n.logger.Error("自动处理同意失败",
					elog.FieldErr(err),
					elog.Any("instId", t.ProcInstID),
					elog.Any("taskId", t.TaskID),
				)
			}
			return
		}

		err := n.engineSvc.Pass(ctx, t.TaskID, "自处理抄送节点审批")
		if err != nil {
			n.logger.Error("自动处理同意失败",
				elog.FieldErr(err),
				elog.Any("instId", t.ProcInstID),
				elog.Any("taskId", t.TaskID),
			)
		}
	}
	return
}

// getUsers 获取需要通知的用户信息
func (n *UserNotification) getUsers(ctx context.Context, tasks []model.Task) ([]user.User, error) {
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
func (n *UserNotification) getRules(ctx context.Context, order order.Order) ([]rule.Rule, error) {
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

func (n *UserNotification) getProperty(nodes []easyflow.Node, currentNodeId string) (easyflow.UserProperty, error) {
	for _, node := range nodes {
		if node.ID == currentNodeId {
			return easyflow.ToNodeProperty[easyflow.UserProperty](node)
		}
	}

	return easyflow.UserProperty{}, nil
}

// 当自动化节点返回信息在流程结束后通知用户，组合所有自动化节点返回的数据，进行消息通知
// 但是全局消息通知关闭的情况下，不会运行此部分
func (n *UserNotification) wantAllResult(ctx context.Context, instanceId int, nodes []easyflow.Node) (map[string]interface{}, error) {
	mergedResult := make(map[string]interface{})
	for _, node := range nodes {
		switch node.Type {
		case "automation":
			property, _ := easyflow.ToNodeProperty[easyflow.AutomationProperty](node)
			// 判断是否进行通知
			if !property.IsNotify || property.NotifyMethod != ProcessEndSend {
				continue
			}

			// 查找自动化任务返回
			result, err := n.taskSvc.FindTaskResult(ctx, instanceId, node.ID)
			if err != nil {
				return nil, err
			}

			// 返回为空则不处理
			if result.WantResult == "" {
				continue
			}

			var wantResult map[string]interface{}
			err = json.Unmarshal([]byte(result.WantResult), &wantResult)
			if err != nil {
				return nil, err
			}

			for key, value := range wantResult {
				mergedResult[key] = value
			}
		}
	}

	if len(mergedResult) == 0 {
		return nil, nil
	}

	return mergedResult, nil
}
