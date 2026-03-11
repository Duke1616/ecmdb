package strategy

import (
	"context"
	"fmt"
	"regexp"
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
	"github.com/Duke1616/enotify/notify/feishu"
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

// chatContext 封装发送所需的所有元数据
type chatContext struct {
	*NotificationData                            // 基础元数据 (Rules, StartUser, TName, WantResult)
	property          easyflow.ChatGroupProperty // 节点配置属性
	members           []user.User                // 运行解析出的成员
	userInputs        []order.FormValue          // 动态加载的用户输入
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
	sendCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
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
func (n *ChatNotification) asyncHandleChat(ctx context.Context, info Info, data *chatContext) {
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

func (n *ChatNotification) sendChatNotifications(ctx context.Context, info Info, data *chatContext) {
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
func (n *ChatNotification) resolveRecipients(ctx context.Context, info Info, data *chatContext) ([]recipient, error) {
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
func (n *ChatNotification) handleExistingGroups(ctx context.Context, info Info, data *chatContext) ([]recipient, error) {
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
func (n *ChatNotification) handleCreateGroup(ctx context.Context, info Info, data *chatContext) ([]recipient, error) {
	// 解析动态群组名称，解析失败则使用默认名称
	chatName := n.resolveChatName(data.property.Create.Name, info, data)

	// 获取团队默认群组
	chats, err := n.teamSvc.GetDefaultChatGroups(ctx, &teamv1.GetDefaultChatGroupsRequest{})
	if err != nil {
		return nil, err
	}

	// 如果群聊存在则不需要多余创建
	chat, ok := slice.Find(chats.Groups, func(src *teamv1.ChatGroup) bool {
		return src.Name == chatName && src.Channel == notificationv1.Channel_LARK_CARD
	})

	if ok {
		// 如果配置了动态成员，同步添加到现有群组
		n.syncMembersToExistingChat(ctx, chat.ChatId, data)
		return []recipient{{
			chatID:  chat.ChatId,
			channel: n.GetChatChannel(chat.Channel.String()),
		}}, nil
	}

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

func (n *ChatNotification) syncMembersToExistingChat(ctx context.Context, chatID string, data *chatContext) {
	members := n.extractMemberIDs(data.members)
	if len(members) == 0 {
		return
	}

	if err := n.addMembersToChat(ctx, chatID, members); err != nil {
		n.Logger.Warn("ChatNotification 添加成员到现有群组失败",
			elog.FieldErr(err),
			elog.String("chatId", chatID),
		)
	}
}

// buildNotifications 组装最终的批量通知对象
func (n *ChatNotification) buildNotifications(info Info, data *chatContext, recipients []recipient) []notification.Notification {
	title := n.resolveTitle(data.property.Title, info, data)
	fields := n.resolveFields(info, data)

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
				Name:     LarkTemplateChatGroup,
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
func (n *ChatNotification) createChatGroup(ctx context.Context, chatName string, data *chatContext) (string, error) {
	memberIDs := n.extractMemberIDs(data.members)
	if len(memberIDs) == 0 {
		return "", fmt.Errorf("未找到有效的群成员")
	}

	req := larkim.NewCreateChatReqBuilder().
		UserIdType(feishu.ReceiveIDTypeUserID).
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
func (n *ChatNotification) fetchChatData(ctx context.Context, info Info) (*chatContext, error) {
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

	return &chatContext{
		NotificationData: base,
		property:         property,
		userInputs:       inputs,
		members:          n.resolveMembers(ctx, info, property),
	}, nil
}

var variableRegex = regexp.MustCompile(`{{(.*?)}}`)

func (n *ChatNotification) resolveTitle(rule string, info Info, data *chatContext) string {
	return n.resolveDynamicString(rule, "{{creator}}发起的{{template}}执行结果", info, data)
}

// resolveChatName 解析动态群组名称
func (n *ChatNotification) resolveChatName(rule string, info Info, data *chatContext) string {
	return n.resolveDynamicString(rule, fmt.Sprintf("【ECMDB】- %s", data.TName), info, data)
}

// resolveDynamicString 解析动态字符串 (支持变量替换)
func (n *ChatNotification) resolveDynamicString(value, defaultVal string, info Info, data *chatContext) string {
	target := value
	if target == "" {
		target = defaultVal
	}

	// 1. 准备变量池
	vars := map[string]string{
		"ticket_id": fmt.Sprintf("%d", info.Order.Id),
		"template":  data.TName,
		"creator":   data.StartUser.DisplayName,
	}

	// 2. 注入表单字段
	for k, v := range info.Order.Data {
		vars["field."+k] = fmt.Sprintf("%v", v)
	}

	// 3. 执行替换
	result := variableRegex.ReplaceAllStringFunc(target, func(match string) string {
		// 提取花括号中的变量名
		res := variableRegex.FindStringSubmatch(match)
		if len(res) < 2 {
			return match
		}

		key := res[1]
		if val, ok := vars[key]; ok {
			return val
		}
		// 如果变量不存在，保持原样
		return match
	})

	// 4. 兜底检查
	if result == "" {
		return defaultVal
	}

	return result
}

// addMembersToChat 为现有群组添加成员
func (n *ChatNotification) addMembersToChat(ctx context.Context, chatID string, memberIDs []string) error {
	req := larkim.NewCreateChatMembersReqBuilder().
		ChatId(chatID).
		MemberIdType(feishu.ReceiveIDTypeUserID).
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
// 每个区块有数据时才插入全宽小标题字段（IsShort: false），
// 使飞书卡片形成清晰的分组视觉层次，无需改动模板。
func (n *ChatNotification) resolveFields(info Info, data *chatContext) []notification.Field {
	var fields []notification.Field

	// 为了快速检查是否存在，将前端传来的选项列表转化为 map
	modeSet := make(map[easyflow.OutputMode]bool, len(data.property.OutputMode))
	for _, mode := range data.property.OutputMode {
		modeSet[mode] = true
	}

	// 强制按固定顺序渲染卡片区块：1. 工单信息 -> 2. 用户提交 -> 3. 执行结果

	// 1. 工单信息（对应 OutputTicketData）
	if modeSet[easyflow.OutputTicketData] {
		ruleFields := rule.GetFields(data.Rules, info.Order.Provide.ToUint8(), info.Order.Data)
		modeFields := n.ConvertRuleFields(ruleFields)
		if len(modeFields) > 0 {
			fields = append(fields, sectionHeader("📋 工单信息"))
			fields = append(fields, modeFields...)
		}
	}

	// 2. 用户提交（对应 OutputUserInput）
	if modeSet[easyflow.OutputUserInput] {
		if len(data.userInputs) > 0 {
			fields = append(fields, sectionHeader("✍️ 用户提交"))
			var inputFields []notification.Field
			for _, input := range data.userInputs {
				inputFields = append(inputFields, notification.Field{
					IsShort: true,
					Tag:     "lark_md",
					Content: fmt.Sprintf("**%s:**\n%v", input.Name, input.Value),
				})
			}
			fields = append(fields, notification.AddRowSpacers(inputFields)...)
		}
	}

	// 3. 执行结果（对应 OutputAutoTask）
	if modeSet[easyflow.OutputAutoTask] {
		modeFields := n.BuildWantResultFields(data.WantResult)
		if len(modeFields) > 0 {
			fields = append(fields, sectionHeader("⚙️ 执行结果"))
			fields = append(fields, modeFields...)
		}
	}

	return fields
}

// sectionHeader 生成全宽小标题字段，用于在飞书卡片中分隔不同数据区块
func sectionHeader(title string) notification.Field {
	return notification.Field{
		IsDivider: true,
		Tag:       "lark_md",
		Content:   "**" + title + "**",
	}
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
