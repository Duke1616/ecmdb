package strategy

import (
	"context"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/department"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/event/domain"
	"github.com/Duke1616/ecmdb/internal/event/service/sender"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	templateSvc "github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/enotify/notify/feishu/card"
	"github.com/ecodeclub/ekit/retry"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"golang.org/x/sync/errgroup"
	"time"
)

type UserNotification struct {
	sender        sender.NotificationSender
	engineSvc     engineSvc.Service
	resultSvc     FetcherResult
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
	userSvc user.Service, resultSvc FetcherResult, departMentSvc department.Service,
	sender sender.NotificationSender) (*UserNotification, error) {

	return &UserNotification{
		sender:          sender,
		engineSvc:       engineSvc,
		templateSvc:     templateSvc,
		orderSvc:        orderSvc,
		userSvc:         userSvc,
		departMentSvc:   departMentSvc,
		resultSvc:       resultSvc,
		logger:          elog.DefaultLogger,
		initialInterval: 5 * time.Second,
		maxRetries:      int32(3),
		maxInterval:     15 * time.Second,
	}, nil
}

func (n *UserNotification) isGlobalNotify(wf workflow.Workflow, instanceId int) bool {
	if !wf.IsNotify {
		n.logger.Warn("【用户节点】全局流程控制未开启消息通知能力",
			elog.Any("instId", instanceId),
		)
		return false
	}

	return true
}

func (n *UserNotification) Send(ctx context.Context, notification domain.StrategyInfo) (bool, error) {
	// 获取流程节点 nodes 信息
	nodes, err := unmarshal(notification.WfInfo)
	if err != nil {
		return false, err
	}

	// 获取当前节点信息
	property, err := getProperty[easyflow.UserProperty](nodes, notification.CurrentNode.NodeID)
	if err != nil {
		return false, err
	}

	// 组合获取数据
	errGroup, ctx := errgroup.WithContext(ctx)
	var (
		wantResult map[string]interface{}
		rules      []rule.Rule
		startUser  user.User
		tName      string
	)
	// 获取自动化任务执行结果
	errGroup.Go(func() error {
		wantResult, err = n.wantAllResult(ctx, notification.InstanceId, nodes)
		return err
	})

	// 解析配置
	errGroup.Go(func() error {
		rules, tName, err = n.getRules(ctx, notification.OrderInfo)
		return err
	})

	// 获取工单创建用户
	errGroup.Go(func() error {
		startUser, err = n.userSvc.FindByUsername(ctx, notification.OrderInfo.CreateBy)
		return err
	})

	if err = errGroup.Wait(); err != nil {
		return false, fmt.Errorf("获取组合数据失败: %w", err)
	}

	// 根据规则生成审批用户
	err = n.resolveRule(ctx, notification.InstanceId, property, startUser, notification.OrderInfo,
		notification.CurrentNode)
	if err != nil {
		n.logger.Error("根据模版信息解析规则失败", elog.FieldErr(err),
			elog.Int("instanceId", notification.InstanceId), elog.String("rule", property.Rule.ToString()))
		return false, err
	}

	// 只有当 Event 结束才能正确获取到 TaskId 信息，放到 Go Routine 异步运行, 通过重试机制获取到数据
	//go n.asyncSendNotification(ctx, instanceId, currentNode, property, startUser, nOrder, rules, wantResult, wf)
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
					elog.Any("instId", notification.InstanceId),
				)
				break
			}

			// 获取当前任务流转到的用户
			tasks, err = n.engineSvc.GetTasksByCurrentNodeId(context.Background(), notification.InstanceId,
				notification.CurrentNode.NodeID)
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
		//var messages []notify.NotifierWrap
		title := rule.GenerateTitle(startUser.DisplayName, tName)
		template := FeishuTemplateApprovalName

		// 判断如果是抄送情况
		if property.IsCC {
			template = FeishuTemplateCC
			title = rule.GenerateCCTitle(startUser.DisplayName, tName)

			// 处理自动通过
			go n.ccPass(ctx, tasks)
		}

		// TODO 全局消息通知，如果没开启则全局不发送消息通知
		// 因为用户是通过规则匹配的，需要动态生成，用户抄送节点需要自动通过
		// 假如这个步骤前置执行，会提前退出，导致流程出现不可逆的严重错误
		if ok := n.isGlobalNotify(notification.WfInfo, notification.InstanceId); !ok {
			return
		}

		// 获取需要传递信息
		userMap := analyzeUsers(users)
		fields := rule.GetFields(rules, notification.OrderInfo.Provide.ToUint8(), notification.OrderInfo.Data)
		for field, value := range wantResult {
			fields = append(fields, card.Field{
				IsShort: true,
				Tag:     "lark_md",
				Content: fmt.Sprintf(`**%s:**\n%v`, field, value),
			})
		}

		ns := slice.Map(tasks, func(idx int, src model.Task) domain.Notification {
			receiver, _ := userMap[src.UserID]
			fmt.Println(receiver, "用户ID")
			return domain.Notification{
				Channel:  domain.ChannelFeishuCard,
				Receiver: receiver,
				Template: domain.Template{
					Name:   template,
					Title:  title,
					Fields: fields,
					Values: []card.Value{
						{
							Key:   "order_id",
							Value: notification.OrderInfo.Id,
						},
						{
							Key:   "task_id",
							Value: src.TaskID,
						},
					},
					HideForm: false,
				},
			}
		})

		_, err = n.sender.BatchSend(context.Background(), ns)
		if err != nil {
			n.logger.Warn("发送消息失败",
				elog.FieldErr(err),
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
func (n *UserNotification) getRules(ctx context.Context, order order.Order) ([]rule.Rule, string, error) {
	// 获取模版详情信息
	t, err := n.templateSvc.DetailTemplate(ctx, order.TemplateId)
	if err != nil {
		return nil, "", err
	}

	rules, err := rule.ParseRules(t.Rules)

	if err != nil {
		return nil, "", err
	}

	return rules, t.Name, nil
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
				n.logger.Warn("【用户节点】自动化节点未开启消息通知")
				return mergedResult, nil
			}

			// 判断模式
			if !containsAutoNotifyMethod(property.NotifyMethod, ProcessEndSend) {
				n.logger.Warn("【用户节点】自动化节点未匹配消息通知规则")
				return mergedResult, nil
			}

			// 获取返回值
			wantResult, err := n.resultSvc.FetchResult(ctx, instanceId, node.ID)
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
