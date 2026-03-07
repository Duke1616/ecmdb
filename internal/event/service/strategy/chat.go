package strategy

import (
	"context"
	"fmt"

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

// chatRequiredData 封装发送所需的所有元数据
type chatRequiredData struct {
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

func (n *ChatNotification) Send(ctx context.Context, info Info) (notification.NotificationResponse, error) {
	// 1. 全局通知校验
	if !n.IsGlobalNotify(info.Workflow) {
		return notification.NewSuccessResponse(0, "全局通知已关闭"), nil
	}

	// 2. 获取发送所需的所有元数据
	data, err := n.fetchRequiredData(ctx, info)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeFetchDataFailed), err.Error()), err
	}

	// 3. 根据 Mode 准备接收群组列表 (包含渠道信息)
	type recipient struct {
		chatID  string
		channel notification.Channel
	}
	var recipients []recipient

	switch data.property.Mode {
	case easyflow.ChatGroupUseExisting:
		recipients = slice.Map(data.property.ChatGroups, func(idx int, src easyflow.ChatGroup) recipient {
			ch := n.GetChatChannel(src.Channel)
			// 如果已有群组未配置渠道，则遵循流程全局配置
			if src.Channel == "" {
				ch = info.Channel
			}
			return recipient{chatID: src.ChatID, channel: ch}
		})
	case easyflow.ChatGroupCreate:
		// 创建新群组，渠道遵循流程全局配置
		chatID, createErr := n.createChatGroup(ctx, data)
		if createErr != nil {
			n.Logger.Error("创建通知群组失败", elog.FieldErr(createErr))
			return notification.NewErrorResponse(string(errs.ErrorCodeFetchDataFailed), createErr.Error()), createErr
		}

		// 异步将新创建的群组绑定到团队
		if data.property.TeamID > 0 {
			go func() {
				title := fmt.Sprintf("【ECmdb】%s-%s", data.startUser.DisplayName, data.template)
				_, bErr := n.teamSvc.BindChatGroup(context.Background(), &teamv1.BindChatGroupRequest{
					Group: &teamv1.ChatGroup{
						TeamId:  data.property.TeamID,
						Name:    title,
						ChatId:  chatID,
						Channel: notificationv1.Channel_LARK_CARD, // 飞书群默认为 LarkCard
					},
				})
				if bErr != nil {
					n.Logger.Error("绑定群组到团队失败", elog.FieldErr(bErr), elog.Int64("teamId", data.property.TeamID))
				}
			}()
		}

		recipients = []recipient{{chatID: chatID, channel: info.Channel}}
	default:
		return notification.NewSuccessResponse(0, "skip: unsupported mode"), nil
	}

	// 3. 构建消息公共内容
	title := rule.GenerateTitle(data.startUser.DisplayName, data.template)
	fields := n.resolveFields(data)

	// 4. 准备批量发送任务
	ns := slice.FilterMap(recipients, func(idx int, r recipient) (notification.Notification, bool) {
		if r.chatID == "" {
			return notification.Notification{}, false
		}

		return notification.Notification{
			Channel:      r.channel,
			ReceiverType: feishu.ReceiveIDTypeChatID,
			WorkFlowID:   info.Workflow.Id,
			Receiver:     r.chatID,
			Template: notification.Template{
				Name:     LarkTemplateCC,
				Title:    title,
				Fields:   fields,
				HideForm: true,
				Values: []notification.Value{
					{
						Key:   "order_id",
						Value: info.Order.Id,
					},
				},
			},
		}, true
	})

	// 5. 调用批量发送
	if len(ns) > 0 {
		if resp, err := n.sender.BatchSend(ctx, ns); err != nil {
			n.Logger.Error("群组通知发送失败", elog.FieldErr(err))
			return resp, err
		}
	}

	return notification.NewSuccessResponse(0, "success"), nil
}

// createChatGroup 创建飞书群组
func (n *ChatNotification) createChatGroup(ctx context.Context, data *chatRequiredData) (string, error) {
	// 1. 整理成员 Feishu UserID
	memberIDs := slice.FilterMap(data.members, func(idx int, src user.User) (string, bool) {
		return src.FeishuInfo.UserId, src.FeishuInfo.UserId != ""
	})

	if len(memberIDs) == 0 {
		return "", fmt.Errorf("未找到有效的群成员")
	}

	// 2. 调用 Lark API 创建群组
	groupName := fmt.Sprintf("【ECmdb】%s-%s", data.startUser.DisplayName, data.template)
	req := larkim.NewCreateChatReqBuilder().
		Body(larkim.NewCreateChatReqBodyBuilder().
			Name(groupName).
			UserIdList(memberIDs).
			Build()).
		Build()

	resp, err := n.larkClient.Im.Chat.Create(ctx, req)
	if err != nil {
		return "", err
	}

	if !resp.Success() {
		return "", fmt.Errorf("lark create chat failed: %s", resp.Msg)
	}

	return *resp.Data.ChatId, nil
}

// fetchRequiredData 并行或顺序获取所有依赖数据
func (n *ChatNotification) fetchRequiredData(ctx context.Context, info Info) (*chatRequiredData, error) {
	// 1. 获取节点属性
	nodes, rawProps, err := n.GetNodeProperty(info, info.CurrentNode.NodeID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrNodeNotConfigured, err)
	}
	property, err := easyflow.ToNodeProperty[easyflow.ChatGroupProperty](easyflow.Node{Properties: rawProps})
	if err != nil {
		return nil, err
	}

	// 2. 获取基础元数据（并行获取）
	data, err := n.FetchRequiredData(ctx, info, nodes)
	if err != nil {
		return nil, err
	}

	// 3. 按需获取扩展数据：用户输入
	var userInputs []order.FormValue
	if slice.Contains(property.OutputMode, easyflow.OutputUserInput) {
		userInputs, _ = n.OrderSvc.FindTaskFormsByOrderID(ctx, info.Order.Id)
	}

	// 4. 按需获取扩展数据：群成员解析
	var members []user.User
	if property.Mode == easyflow.ChatGroupCreate {
		targets := n.EnrichTargets(info, property.Assignees)
		members, _ = n.assigneeService.Resolve(ctx, targets)
	}

	return &chatRequiredData{
		property:   property,
		startUser:  data.StartUser,
		rules:      data.Rules,
		template:   data.TName,
		wantResult: data.WantResult,
		userInputs: userInputs,
		members:    members,
		orderData:  info.Order.Data,
		provide:    info.Order.Provide.ToUint8(),
	}, nil
}

// resolveFields 根据输出模式解析展示字段
func (n *ChatNotification) resolveFields(data *chatRequiredData) []notification.Field {
	var fields []notification.Field

	for _, mode := range data.property.OutputMode {
		switch mode {
		case easyflow.OutputTicketData:
			ruleFields := rule.GetFields(data.rules, data.provide, data.orderData)
			fields = append(fields, n.ConvertRuleFields(ruleFields)...)
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
