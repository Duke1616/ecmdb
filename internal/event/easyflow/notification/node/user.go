package node

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/department"
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
	integrations  []method.NotifyIntegration
	engineSvc     engineSvc.Service
	taskSvc       task.Service
	userSvc       user.Service
	departMentSvc department.Service
	templateSvc   templateSvc.Service
	orderSvc      order.Service

	initialInterval time.Duration
	maxInterval     time.Duration
	maxRetries      int32
	logger          *elog.Component
}

func NewUserNotification(engineSvc engineSvc.Service, templateSvc templateSvc.Service, orderSvc order.Service,
	userSvc user.Service, taskSvc task.Service, departMentSvc department.Service,
	integrations []method.NotifyIntegration) (*UserNotification, error) {

	return &UserNotification{
		engineSvc:       engineSvc,
		templateSvc:     templateSvc,
		orderSvc:        orderSvc,
		userSvc:         userSvc,
		departMentSvc:   departMentSvc,
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
		n.logger.Warn("【用户节点】全局流程控制未开启消息通知能力",
			elog.Any("instId", instanceId),
		)
		return false
	}

	return true
}

func (n *UserNotification) Send(ctx context.Context, nOrder order.Order, wf workflow.Workflow,
	instanceId int, currentNode *model.Node) (bool, error) {
	// 判断是否开启
	if ok := n.isNotify(wf, instanceId); !ok {
		return false, nil
	}

	// 获取流程节点 nodes 信息
	nodes, err := unmarshal(wf)
	if err != nil {
		return false, err
	}

	// 获取当前节点信息
	property, err := getProperty[easyflow.UserProperty](nodes, currentNode.NodeID)
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

	// 获取工单创建用户
	startUser, err := n.userSvc.FindByUsername(ctx, nOrder.CreateBy)
	if err != nil {
		return false, err
	}

	// 根据规则生成审批用户
	err = n.resolveRule(ctx, instanceId, property, startUser, nOrder, currentNode)
	if err != nil {
		n.logger.Error("根据模版信息解析规则失败", elog.FieldErr(err),
			elog.Int("instanceId", instanceId), elog.String("rule", property.Rule.ToString()))
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
			tasks, err = n.engineSvc.GetTasksByCurrentNodeId(context.Background(), instanceId, currentNode.NodeID)
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
				messages = integration.Notifier.Builder(title, users, template, method.NewNotifyParamsBuilder().
					SetRules(rules).
					SetOrder(nOrder).
					SetTasks(tasks).
					SetWantResult(wantResult).
					Build())
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

// 当自动化节点返回信息在流程结束后通知用户，组合所有自动化节点返回的数据，进行消息通知
// 但是全局消息通知关闭的情况下，不会运行此部分
func (n *UserNotification) wantAllResult(ctx context.Context, instanceId int, nodes []easyflow.Node) (map[string]interface{}, error) {
	mergedResult := make(map[string]interface{})
	for _, node := range nodes {
		switch node.Type {
		case "automation":
			property, _ := easyflow.ToNodeProperty[easyflow.AutomationProperty](node)
			// 判断是否开启消息发送，以及是否为立即发送
			if !property.IsNotify {
				return nil, fmt.Errorf("【用户节点】自动化节点未开启消息通知")
			}

			if !containsAutoNotifyMethod(property.NotifyMethod, ProcessEndSend) {
				return nil, fmt.Errorf("【用户节点】自动化节点未匹配消息通知规则")
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

	return mergedResult, nil
}

func (n *UserNotification) resolveRule(ctx context.Context, instanceId int, userProperty easyflow.UserProperty,
	startUser user.User, nOrder order.Order,
	currentNode *model.Node) error {
	switch userProperty.Rule {
	case easyflow.LEADER:
		defaultUserIds := currentNode.UserIDs
		currentNode.UserIDs = []string{}
		depart, err := n.resolveDepartment(ctx, instanceId, startUser)
		if err != nil {
			return err
		}

		users, err := n.userSvc.FindByIds(ctx, depart.Leaders)
		if err != nil {
			return err
		}

		if len(users) == 0 {
			currentNode.UserIDs = append(currentNode.UserIDs, defaultUserIds...)
			return nil
		}

		for _, u := range users {
			currentNode.UserIDs = append(currentNode.UserIDs, u.Username)
		}
	case easyflow.MAIN_LEADER:
		defaultUserIds := currentNode.UserIDs
		currentNode.UserIDs = []string{}

		depart, err := n.resolveDepartment(ctx, instanceId, startUser)
		if err != nil {
			return err
		}

		u, err := n.userSvc.FindById(ctx, depart.MainLeader)
		if err != nil {
			return err
		}

		if u.Id == 0 {
			currentNode.UserIDs = append(currentNode.UserIDs, defaultUserIds...)
			return nil
		}

		currentNode.UserIDs = append(currentNode.UserIDs, u.Username)
	case easyflow.TEMPLATE:
		value, ok := nOrder.Data[userProperty.TemplateField]
		if !ok {
			return fmt.Errorf("根据模版字段查询失败，不存在")
		}
		// 处理字符串情况
		if str, isString := value.(string); isString {
			u, err := n.userSvc.FindByUsername(ctx, str)
			if err != nil {
				return err // 处理错误
			}
			currentNode.UserIDs = append(currentNode.UserIDs, u.Username)

			return nil
		}

		// 处理数据情况
		if arr, isArray := value.([]interface{}); isArray {
			for _, item := range arr {
				if str, isString := item.(string); isString {
					u, err := n.userSvc.FindByUsername(ctx, str)
					if err != nil {
						return err // 处理错误
					}
					currentNode.UserIDs = append(currentNode.UserIDs, u.Username)
				}
			}

			return nil
		}

		return fmt.Errorf("未匹配任何模版成功")
	case easyflow.FOUNDER:
		currentNode.UserIDs = append(currentNode.UserIDs, startUser.Username)
	case easyflow.APPOINT:
		if currentNode.UserIDs == nil || len(currentNode.UserIDs) == 0 {
			// TODO 后续处理、如果触发这条线路，应该做错误消息提醒
			n.logger.Error("没有指定的审批人，系统将自动插入流程管理员用户，防止流程中断报错")
		}
	}

	return nil
}

func (n *UserNotification) resolveDepartment(ctx context.Context, instanceId int, user user.User) (
	department.Department, error) {
	// 判断如果所属组不为空
	if user.DepartmentId == 0 {
		return department.Department{}, fmt.Errorf("用户所属组为空")
	}

	depart, err := n.departMentSvc.FindById(ctx, user.DepartmentId)
	if err != nil {
		return department.Department{}, err
	}

	return depart, nil
}
