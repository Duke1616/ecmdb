package strategy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	notificationv1 "github.com/Duke1616/ecmdb/api/proto/gen/notification/v1"
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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/structpb"
)

type UserNotification struct {
	sender          sender.NotificationSender
	engineSvc       engineSvc.Service
	resultSvc       FetcherResult
	userSvc         user.Service
	departmentSvc   department.Service
	templateSvc     templateSvc.Service
	orderSvc        order.Service
	notificationSvc notificationv1.NotificationServiceClient

	initialInterval time.Duration
	maxInterval     time.Duration
	maxRetries      int32
	logger          *elog.Component
}

func NewUserNotification(engineSvc engineSvc.Service, templateSvc templateSvc.Service, orderSvc order.Service,
	userSvc user.Service, resultSvc FetcherResult, departmentSvc department.Service,
	sender sender.NotificationSender, notificationSvc notificationv1.NotificationServiceClient) (*UserNotification, error) {

	return &UserNotification{
		sender:          sender,
		engineSvc:       engineSvc,
		templateSvc:     templateSvc,
		orderSvc:        orderSvc,
		userSvc:         userSvc,
		departmentSvc:   departmentSvc,
		resultSvc:       resultSvc,
		notificationSvc: notificationSvc,
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
	// 获取流程节点信息
	nodes, err := unmarshal(notification.WfInfo)
	if err != nil {
		return false, fmt.Errorf("解析流程信息失败: %w", err)
	}

	// 获取当前节点属性
	property, err := getProperty[easyflow.UserProperty](nodes, notification.CurrentNode.NodeID)
	if err != nil {
		return false, fmt.Errorf("获取节点属性失败: %w", err)
	}

	// 并行获取所需数据
	data, err := n.fetchRequiredData(ctx, notification, nodes)
	if err != nil {
		return false, fmt.Errorf("获取组合数据失败: %w", err)
	}

	// 为异步发送设置一个独立的超时 ctx，避免无限挂起
	sendCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)

	// 根据规则生成审批用户
	if err = n.resolveRule(sendCtx, notification.InstanceId, property, data.startUser,
		notification.OrderInfo, notification.CurrentNode); err != nil {
		cancel() // 立即取消，因为发生了错误
		n.logger.Error("解析规则失败",
			elog.FieldErr(err),
			elog.Int("instanceId", notification.InstanceId),
			elog.String("rule", property.Rule.ToString()))
		return false, fmt.Errorf("根据规则解析用户失败：%w", err)
	}

	// 异步发送通知，将 cancel 函数传给异步任务
	// 只有当 Event 结束才能正确获取到 TaskId 信息，放到 Go Routine 异步运行, 通过重试机制获取到数据
	go func() {
		defer func() {
			cancel()
			if r := recover(); r != nil {
				n.logger.Error("异步发送通知发生panic", elog.Any("recover", r))
			}
		}()
		n.asyncSendNotification(sendCtx, notification, property, data)
	}()

	return true, nil
}

func (n *UserNotification) asyncSendNotification(ctx context.Context, notification domain.StrategyInfo,
	property easyflow.UserProperty, data *notificationData) {
	// 1. 获取任务信息（带重试）
	tasks, err := n.fetchTasksWithRetry(ctx, notification)
	if err != nil {
		n.logger.Error("获取任务信息失败",
			elog.FieldErr(err),
			elog.Int("instanceId", notification.InstanceId))
		return
	}

	// 2. 生成消息数据（先生成默认的）
	title := rule.GenerateTitle(data.startUser.DisplayName, data.tName)
	template := FeishuTemplateApprovalName

	// 3. 判断是否是抄送情况，如果是，则修改模板和标题，且先执行自动通过
	if property.IsCC {
		template = FeishuTemplateCC
		title = rule.GenerateCCTitle(data.startUser.DisplayName, data.tName)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					n.logger.Error("ccPass协程异常恢复", elog.Any("panic", r))
				}
			}()
			n.ccPass(ctx, tasks)
		}()
	}

	// 4. 判断是否全局允许通知，若不允许则返回（确保自动通过已执行）
	// 如果节点是抄送节点需要自动结束该节点，如果这块代码放到抄送逻辑前，会影响代码处理
	if ok := n.isGlobalNotify(notification.WfInfo, notification.InstanceId); !ok {
		return
	}

	// 5. 获取用户详情信息
	users, err := n.getUsers(ctx, tasks)
	if err != nil {
		n.logger.Error("用户查询失败", elog.FieldErr(err))
	}
	userMap := analyzeUsers(users)

	// 6、如果消息来源是告警转工单，则通过告警模版发送消息
	if notification.OrderInfo.Provide.IsAlert() {
		if err = n.send(ctx, tasks, notification.OrderInfo, userMap); err != nil {
			n.logger.Error("告警转工单，消息发送失败", elog.FieldErr(err))
		}

		return
	}

	// 7、. 获取需要传递的信息字段
	fields := rule.GetFields(data.rules, notification.OrderInfo.Provide.ToUint8(), notification.OrderInfo.Data)
	for field, value := range data.wantResult {
		fields = append(fields, card.Field{
			IsShort: true,
			Tag:     "lark_md",
			Content: fmt.Sprintf(`**%s:**\n%v`, field, value),
		})
	}

	ns := slice.Map(tasks, func(idx int, src model.Task) domain.Notification {
		receiver, _ := userMap[src.UserID]
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

	if _, err = n.sender.BatchSend(ctx, ns); err != nil {
		n.logger.Warn("发送消息失败",
			elog.FieldErr(err),
		)
		return
	}
}

func (n *UserNotification) send(ctx context.Context, tasks []model.Task,
	oInfo order.Order, userMap map[string]string) error {
	userIds := slice.Map(tasks, func(idx int, src model.Task) string {
		receiver, _ := userMap[src.UserID]
		return receiver
	})

	// 组合消息
	var tParams map[string]any
	tParams, err := toPureGoType(oInfo.NotificationConf.TemplateParams)
	if err != nil {
		return fmt.Errorf("解析模版数据失败: %w", err)
	}

	params, err := structpb.NewStruct(tParams)
	if err != nil {
		return fmt.Errorf("解析模版数据失败: %w", err)
	}

	// 发送消息
	_, err = n.notificationSvc.SendNotification(ctx, &notificationv1.SendNotificationRequest{
		Notification: &notificationv1.Notification{
			Key:            "ecmdb",
			Receivers:      userIds,
			Channel:        order.ChannelToDomainProto(oInfo.NotificationConf.Channel),
			TemplateParams: params,
			TemplateId:     oInfo.NotificationConf.TemplateID,
		},
	})

	if err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	return nil
}

// 分解出的辅助方法
func (n *UserNotification) fetchTasksWithRetry(ctx context.Context, notification domain.StrategyInfo) ([]model.Task, error) {
	strategy, err := retry.NewExponentialBackoffRetryStrategy(n.initialInterval, n.maxInterval, n.maxRetries)
	if err != nil {
		return nil, err
	}

	var tasks []model.Task
	for {
		d, ok := strategy.Next()
		if !ok {
			return nil, fmt.Errorf("处理执行任务超过最大重试次数")
		}

		tasks, err = n.engineSvc.GetTasksByCurrentNodeId(ctx, notification.InstanceId, notification.CurrentNode.NodeID)
		if err == nil && len(tasks) > 0 {
			return tasks, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(d):
			continue
		}
	}
}

// 提取的数据获取逻辑
func (n *UserNotification) fetchRequiredData(ctx context.Context, notification domain.StrategyInfo,
	nodes []easyflow.Node) (*notificationData, error) {
	var data notificationData
	errGroup, ctx := errgroup.WithContext(ctx)

	// 获取自动化任务执行结果
	errGroup.Go(func() error {
		var err error
		data.wantResult, err = n.wantAllResult(ctx, notification.InstanceId, nodes)
		return err
	})

	// 解析配置
	errGroup.Go(func() error {
		var err error
		data.rules, data.tName, err = n.getRules(ctx, notification.OrderInfo)
		return err
	})

	// 获取工单创建用户
	errGroup.Go(func() error {
		var err error
		data.startUser, err = n.userSvc.FindByUsername(ctx, notification.OrderInfo.CreateBy)
		return err
	})

	if err := errGroup.Wait(); err != nil {
		return nil, err
	}

	return &data, nil
}

func toPureGoType(input map[string]any) (map[string]any, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var pureMap map[string]interface{}
	err = json.Unmarshal(b, &pureMap)
	if err != nil {
		return nil, err
	}

	return pureMap, nil
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

		users, err := n.userSvc.FindByUsernames(ctx, depart.Leaders)
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

		u, err := n.userSvc.FindByUsername(ctx, depart.MainLeader)
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
		switch v := value.(type) {
		// 处理字符串情况
		case string:
			u, err := n.userSvc.FindByUsername(ctx, v)
			if err != nil {
				return fmt.Errorf("failed to find user '%s': %w", v, err)
			}
			currentNode.UserIDs = append(currentNode.UserIDs, u.Username)
			return nil
		// 处理数组情况
		case []interface{}:
			return n.processUserArray(ctx, v, currentNode)
		// 处理 MongoDB 的 BSON 数组类型
		case primitive.A:
			return n.processUserArray(ctx, v, currentNode)
		default:
			return fmt.Errorf("unexpected type %T for template field, expected string or []string", value)
		}
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

func (n *UserNotification) processUserArray(ctx context.Context, arr []interface{}, currentNode *model.Node) error {
	for _, item := range arr {
		if str, ok := item.(string); ok {
			u, err := n.userSvc.FindByUsername(ctx, str)
			if err != nil {
				return fmt.Errorf("failed to find user '%s': %w", str, err)
			}
			currentNode.UserIDs = append(currentNode.UserIDs, u.Username)
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

	depart, err := n.departmentSvc.FindById(ctx, user.DepartmentId)
	if err != nil {
		return department.Department{}, err
	}

	return depart, nil
}

// 辅助结构体
type notificationData struct {
	wantResult map[string]interface{}
	rules      []rule.Rule
	startUser  user.User
	tName      string
}
