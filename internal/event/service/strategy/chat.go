package strategy

import (
	"context"
	"fmt"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	notificationv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/notification/v1"
	teamv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/team"
	"github.com/Duke1616/ecmdb/internal/event/errs"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/sender"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/ecmdb/pkg/resolve"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type ChatNotification struct {
	BaseStrategy
	sender          sender.NotificationSender
	larkClient      *lark.Client
	assigneeService *resolve.Engine
	teamSvc         teamv1.TeamServiceClient
}

func NewChatNotification(base BaseStrategy, sender sender.NotificationSender,
	larkClient *lark.Client, assigneeService *resolve.Engine, teamSvc teamv1.TeamServiceClient) *ChatNotification {
	return &ChatNotification{
		BaseStrategy:    base,
		sender:          sender,
		larkClient:      larkClient,
		assigneeService: assigneeService,
		teamSvc:         teamSvc,
	}
}

// chatData 封装发送所需的所有元数据
type chatData struct {
	property   easyflow.ChatGroupProperty
	startUser  user.User
	rules      []rule.Rule
	template   string
	wantResult map[string]interface{}
	userInputs []order.FormValue
	members    []user.User
	orderData  map[string]interface{}
	provide    uint8
}

// recipient 内部接收者描述
type recipient struct {
	chatID  string
	channel notification.Channel
}

func (n *ChatNotification) Send(ctx context.Context, info Info) (notification.NotificationResponse, error) {
	// 1. 获取通知元数据 (同步获取以注入 UserIDs)
	data, err := n.fetchChatData(ctx, info)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeFetchDataFailed), err.Error()), err
	}

	// 2. 注入审批人 (利用 data.members 确保引擎创建任务，以便后续自动推进)
	info.CurrentNode.UserIDs = slice.Map(data.members, func(idx int, u user.User) string {
		return u.Username
	})

	// 3. 异步处理：等待任务创建、发送群组消息、自动推进流程
	sendCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	go func() {
		defer cancel()
		defer func() {
			if r := recover(); r != nil {
				n.Logger.Error("ChatNotification async panic", elog.Any("recover", r))
			}
		}()
		n.asyncHandleChat(sendCtx, info, data)
	}()

	return notification.NotificationResponse{}, nil
}

// asyncHandleChat 异步处理核心逻辑
func (n *ChatNotification) asyncHandleChat(ctx context.Context, info Info, data *chatData) {
	n.Logger.Info("ChatNotification 开始异步处理群组通知",
		elog.Int("instId", info.InstID),
		elog.String("node", info.CurrentNode.NodeID))

	// 1. 获取任务信息 (带重试，等待引擎完成任务下发)
	tasks, err := n.FetchTasksWithRetry(ctx, info)
	if err != nil {
		n.Logger.Error("ChatNotification 获取任务失败", elog.FieldErr(err), elog.Int("instId", info.InstID))
		return
	}

	// 2. 检查全局消息通知开关
	if !n.IsGlobalNotify(info.Workflow) {
		n.Logger.Info("ChatNotification 全局流程通知开关已关闭，跳过消息发送，仅执行流程自动推进",
			elog.Int64("wfId", info.Workflow.Id))
	} else {
		n.sendChatNotifications(ctx, info, data)
	}

	// 3. 自动推进流程
	n.autoPassTasks(ctx, tasks)
}

func (n *ChatNotification) autoPassTasks(ctx context.Context, tasks []model.Task) {
	for _, t := range tasks {
		if err := n.EngineSvc.Pass(ctx, t.TaskID, "ChatGroup Auto Pass"); err != nil {
			n.Logger.Error("ChatNotification 流程自动推进失败",
				elog.FieldErr(err),
				elog.Int("taskId", t.TaskID))
			continue
		}

		n.Logger.Info("ChatNotification 节点任务已自动推进",
			elog.Int("taskId", t.TaskID))

		// 抄送或自动通知通常一人通过全组通过 (非会签)
		if t.IsCosigned != 1 {
			return
		}
	}
}

func (n *ChatNotification) sendChatNotifications(ctx context.Context, info Info, data *chatData) {
	recipients, err := n.resolveRecipients(ctx, info, data)
	if err != nil {
		n.Logger.Error("ChatNotification 解析接收群组失败", elog.FieldErr(err))
		return
	}

	if len(recipients) == 0 {
		n.Logger.Info("ChatNotification 未匹配到合法的群组接收者，跳过发送")
		return
	}

	notifications := n.buildNotifications(info, data, recipients)

	if _, err = n.sender.BatchSend(ctx, notifications); err != nil {
		n.Logger.Warn("ChatNotification 发送群组消息失败", elog.FieldErr(err))
		return
	}

	n.Logger.Info("ChatNotification 群组消息批量发送成功",
		elog.Int("recipientCount", len(recipients)))
}

// resolveRecipients 根据模式分发解析接收者逻辑
func (n *ChatNotification) resolveRecipients(ctx context.Context, info Info, data *chatData) ([]recipient, error) {
	switch data.property.Mode {
	case easyflow.ChatGroupUseExisting:
		return n.handleExistingGroups(ctx, info, data)
	case easyflow.ChatGroupCreate:
		return n.handleCreateGroup(ctx, info, data)
	default:
		n.Logger.Warn("未知的群组通知模式", elog.Any("mode", data.property.Mode))
		return nil, nil
	}
}

// handleExistingGroups 处理现有群组逻辑：必须显式指定 ChatGroupIDs
func (n *ChatNotification) handleExistingGroups(ctx context.Context, info Info, data *chatData) ([]recipient, error) {
	// 1. 检查是否配置了群组 ID
	if len(data.property.ChatGroupIDs) == 0 {
		n.Logger.Warn("ChatNotification (ExistingMode) 未显式配置群组 ID，跳过处理",
			elog.Int("instId", info.InstID))
		return nil, nil
	}

	// 2. 获取群组详情
	resp, err := n.teamSvc.GetChatGroupByIds(ctx, &teamv1.GetChatGroupByIdsRequest{Ids: data.property.ChatGroupIDs})
	if err != nil {
		return nil, fmt.Errorf("获取群组详情失败: %w", err)
	}

	// 3. 如果配置了动态成员，同步添加到现有群组
	if members := n.extractMemberIDs(data.members); len(members) > 0 {
		for _, cg := range resp.Groups {
			if err = n.addMembersToChat(ctx, cg.ChatId, members); err != nil {
				n.Logger.Warn("ChatNotification 添加成员到现有群组失败",
					elog.FieldErr(err),
					elog.String("chatId", cg.ChatId))
			}
		}
	}

	// 4. 转换为接收者对象
	return slice.Map(resp.Groups, func(idx int, src *teamv1.ChatGroup) recipient {
		ch := n.GetChatChannel(src.Channel.String())
		if src.Channel == notificationv1.Channel_CHANNEL_UNSPECIFIED {
			ch = info.Channel
		}
		return recipient{chatID: src.ChatId, channel: ch}
	}), nil
}

// handleCreateGroup 处理新建群组及其团队绑定
func (n *ChatNotification) handleCreateGroup(ctx context.Context, info Info, data *chatData) ([]recipient, error) {
	// 群组名称
	chatName := fmt.Sprintf("【ECMDB】- %s", data.template)

	// 创建群组
	chatID, err := n.createChatGroup(ctx, chatName, data)
	if err != nil {
		n.Logger.Error("创建通知群组失败", elog.FieldErr(err))
		return nil, err
	}

	// 异步绑定团队
	go n.asyncBindGroupToTeam(chatName, chatID)

	return []recipient{{chatID: chatID, channel: info.Channel}}, nil
}

// asyncBindGroupToTeam 异步将新群组绑定到发起人的团队
func (n *ChatNotification) asyncBindGroupToTeam(chatName, chatID string) {
	// 如果 Team 指定为 0 ，则代表全局默认团队
	_, err := n.teamSvc.BindChatGroup(context.Background(), &teamv1.BindChatGroupRequest{
		Group: &teamv1.ChatGroup{
			TeamId:  0,
			Name:    chatName,
			ChatId:  chatID,
			Channel: notificationv1.Channel_LARK_CARD,
		},
	})

	if err != nil {
		n.Logger.Error("异步绑定群组到默认团队", elog.FieldErr(err))
	}
}

// buildNotifications 组装最终的批量通知对象
func (n *ChatNotification) buildNotifications(info Info, data *chatData, recipients []recipient) []notification.Notification {
	title := rule.GenerateTitle(data.startUser.DisplayName, data.template)
	fields := n.resolveFields(data)

	return slice.FilterMap(recipients, func(idx int, r recipient) (notification.Notification, bool) {
		if r.chatID == "" {
			return notification.Notification{}, false
		}

		return notification.Notification{
			Channel:      r.channel,
			ReceiverType: notification.ReceiverTypeChatGroup,
			WorkFlowID:   info.Workflow.Id,
			Receiver:     r.chatID,
			Template: notification.Template{
				Name:     LarkTemplateCC,
				Title:    title,
				Fields:   fields,
				HideForm: true,
				Values: []notification.Value{
					{Key: "order_id", Value: info.Order.Id},
				},
			},
		}, true
	})
}

// createChatGroup 创建飞书群组
func (n *ChatNotification) createChatGroup(ctx context.Context, chatName string, data *chatData) (string, error) {
	memberIDs := n.extractMemberIDs(data.members)
	if len(memberIDs) == 0 {
		return "", fmt.Errorf("未找到有效的群成员")
	}

	req := larkim.NewCreateChatReqBuilder().
		UserIdType(`user_id`).
		Body(larkim.NewCreateChatReqBodyBuilder().
			Name(chatName).
			UserIdList(memberIDs).
			Build()).
		Build()

	resp, err := n.larkClient.Im.V1.Chat.Create(ctx, req)
	if err != nil {
		return "", err
	}

	if !resp.Success() {
		return "", fmt.Errorf("飞书创建群组失败: %s", resp.Msg)
	}

	return *resp.Data.ChatId, nil
}

// fetchChatData 获取发送消息所需的全量元数据
func (n *ChatNotification) fetchChatData(ctx context.Context, info Info) (*chatData, error) {
	// 1. 获取节点属性与基础数据
	nodes, rawProps, err := n.GetNodeProperty(info, info.CurrentNode.NodeID)
	if err != nil {
		return nil, err
	}

	property, err := easyflow.ToNodeProperty[easyflow.ChatGroupProperty](easyflow.Node{Properties: rawProps})
	if err != nil {
		return nil, err
	}

	base, err := n.FetchRequiredData(ctx, info, nodes)
	if err != nil {
		return nil, err
	}

	// 2. 依据配置拉取扩展信息
	var inputs []order.FormValue
	if slice.Contains(property.OutputMode, easyflow.OutputUserInput) {
		inputs, _ = n.OrderSvc.FindTaskFormsByOrderID(ctx, info.Order.Id)
	}

	return &chatData{
		property:   property,
		startUser:  base.StartUser,
		rules:      base.Rules,
		template:   base.TName,
		wantResult: base.WantResult,
		userInputs: inputs,
		members:    n.resolveMembers(ctx, info, property),
		orderData:  info.Order.Data,
		provide:    info.Order.Provide.ToUint8(),
	}, nil
}

// addMembersToChat 为现有群组添加成员
func (n *ChatNotification) addMembersToChat(ctx context.Context, chatID string, memberIDs []string) error {
	req := larkim.NewCreateChatMembersReqBuilder().
		ChatId(chatID).
		MemberIdType("user_id").
		Body(larkim.NewCreateChatMembersReqBodyBuilder().
			IdList(memberIDs).
			Build()).
		Build()

	resp, err := n.larkClient.Im.V1.ChatMembers.Create(ctx, req)
	if err != nil {
		return err
	}

	if !resp.Success() {
		// 忽略重复加入的错误 (99991663: 已经在群中; 99991668: 部分已经在群中)
		if resp.Code == 99991663 || resp.Code == 99991668 {
			return nil
		}
		return fmt.Errorf("飞书群组拉人失败[code: %d]: %s", resp.Code, resp.Msg)
	}

	return nil
}

// resolveMembers 解析规则获取最终用户列表
func (n *ChatNotification) resolveMembers(ctx context.Context, info Info, property easyflow.ChatGroupProperty) []user.User {
	// 1. 如果未配置分配规则，则不解析成员
	if len(property.Assignees) == 0 {
		return nil
	}

	// 2. 解析配置的参与者规则
	targets := n.EnrichTargets(info, property.Assignees)
	users, _ := n.assigneeService.Resolve(ctx, targets)
	return users
}

// resolveFields 依据 OutputMode 解析通知内容字段
func (n *ChatNotification) resolveFields(data *chatData) []notification.Field {
	var fields []notification.Field
	for _, mode := range data.property.OutputMode {
		switch mode {
		case easyflow.OutputTicketData:
			fields = append(fields, n.ConvertRuleFields(rule.GetFields(data.rules, data.provide, data.orderData))...)
		case easyflow.OutputAutoTask:
			fields = append(fields, n.BuildWantResultFields(data.wantResult)...)
		case easyflow.OutputUserInput:
			for _, input := range data.userInputs {
				fields = append(fields, notification.Field{
					IsShort: true,
					Tag:     "lark_md",
					Content: fmt.Sprintf("**%s:**\n%v", input.Name, input.Value),
				})
			}
		}
	}
	return fields
}

// extractMemberIDs 提取飞书 ID 列表 (优先使用 UserId, 因为接口显式指定了 user_id 类型)
func (n *ChatNotification) extractMemberIDs(users []user.User) []string {
	return slice.FilterMap(users, func(idx int, src user.User) (string, bool) {
		if src.FeishuInfo.UserId != "" {
			return src.FeishuInfo.UserId, true
		}
		return src.FeishuInfo.UserId, src.FeishuInfo.UserId != ""
	})
}
