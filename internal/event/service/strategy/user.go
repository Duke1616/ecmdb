package strategy

import (
	"context"
	"fmt"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	notificationv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/notification/v1"
	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/event/errs"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/sender"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"google.golang.org/protobuf/types/known/structpb"
)

type UserNotification struct {
	Service
	sender          sender.NotificationSender
	departmentSvc   department.Service
	notificationSvc notificationv1.NotificationServiceClient
}

func NewUserNotification(base Service, departmentSvc department.Service,
	sender sender.NotificationSender, notificationSvc notificationv1.NotificationServiceClient) *UserNotification {

	return &UserNotification{
		Service:         base,
		sender:          sender,
		departmentSvc:   departmentSvc,
		notificationSvc: notificationSvc,
	}
}

func (n *UserNotification) Send(ctx context.Context, info Info) (notification.NotificationResponse, error) {
	// 1. 获取当前节点信息
	nodes, rawProps, err := n.GetNodeProperty(info, info.CurrentNode.NodeID)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), err.Error()), fmt.Errorf("%w: %v", errs.ErrNodeNotConfigured, err)
	}

	property, err := easyflow.ToNodeProperty[easyflow.UserProperty](easyflow.Node{Properties: rawProps})
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), err.Error()), fmt.Errorf("%w: %v", errs.ErrNodeNotConfigured, err)
	}

	// 2. 解析审批人
	users, err := n.ResolveAssignees(ctx, &info, property.NormalizeAssignees())
	if err != nil {
		n.Logger().Error("解析审批人规则失败", elog.FieldErr(err), elog.String("node", info.CurrentNode.NodeID))
		return notification.NewErrorResponse(string(errs.ErrorCodeResolveRuleFailed), err.Error()), err
	}

	// 3. 构建通知元数据
	data, err := n.FetchRequiredData(ctx, info, nodes)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeFetchDataFailed), err.Error()), err
	}

	// 4. 异步处理
	n.SafeGo(ctx, 3*time.Minute, func(sendCtx context.Context) {
		n.asyncSendNotification(sendCtx, info, property, data, NewRecipientMap(users, info.Channel))
	})

	return notification.NewSuccessResponse(0, "success"), nil
}

func (n *UserNotification) asyncSendNotification(ctx context.Context, info Info,
	property easyflow.UserProperty, data *NotificationData, userMap RecipientMap) {
	// 1. 尝试拉取任务，支持重试（引擎异步落库需时间）
	tasks, err := n.FetchTasksWithRetry(ctx, info)
	if err != nil {
		n.Logger().Warn("UserNotification 任务拉取失败，跳过通知发送", elog.FieldErr(err), elog.Int("instanceId", info.InstID))
		return
	}

	// 2. 准备通知元数据（模板、标题、抄送自动处理）
	template, title := n.prepareNotificationMetadata(ctx, property, data, tasks)

	// 3. 全局通知校验（确保抄送自动通过已执行后校验）
	if ok := n.IsGlobalNotify(info.Workflow); !ok {
		return
	}

	// 4. 特殊处理：告警转工单通知
	if info.Order.Provide.IsAlert() {
		if err = n.sendAlertNotification(ctx, tasks, info.Order, userMap); err != nil {
			n.Logger().Error("告警转工单消息发送失败", elog.FieldErr(err))
		}
		return
	}

	// 5. 发行标准通知卡片
	fields := n.PrepareCommonFields(info, data)
	notifications := n.buildNotifications(info, property, tasks, userMap, template, title, fields)
	for _, msg := range notifications {
		if _, err = n.sender.Send(ctx, msg); err != nil {
			n.Logger().Error("UserNotification 消息发送失败", elog.FieldErr(err), elog.String("receiver", msg.Receiver))
		}
	}
}

func (n *UserNotification) prepareNotificationMetadata(ctx context.Context, property easyflow.UserProperty,
	data *NotificationData, tasks []model.Task) (template workflow.NotifyType, title string) {
	title = rule.GenerateTitle(data.StartUser.DisplayName, data.TName)
	template = LarkTemplateApprovalName
	return
}

func (n *UserNotification) buildNotifications(info Info, property easyflow.UserProperty, tasks []model.Task,
	userMap RecipientMap, template workflow.NotifyType, title string, fields []notification.Field) []notification.Notification {
	return slice.Map(tasks, func(idx int, src model.Task) notification.Notification {
		receiver := userMap.GetID(src.UserID)
		return notification.Notification{
			Channel:    info.Channel,
			WorkFlowID: info.Workflow.Id,
			Receiver:   receiver,
			Template: notification.Template{
				Name:   template,
				Title:  title,
				Fields: fields,
				Values: []notification.Value{
					{Key: "order_id", Value: info.Order.Id},
					{Key: "task_id", Value: src.TaskID},
				},
				HideForm: false,
				InputFields: slice.Map(property.Fields, func(idx int, src easyflow.Field) notification.InputField {
					return notification.InputField{
						Name:     src.Name,
						Key:      src.Key,
						Type:     notification.FieldType(src.Type),
						Required: src.Required,
						Value:    src.Value,
						ReadOnly: src.ReadOnly,
						Options: slice.Map(src.Options, func(idx int, src easyflow.Option) notification.InputOption {
							return notification.InputOption{
								Label: src.Label,
								Value: src.Value,
							}
						}),
						Props: src.Props,
					}
				}),
			},
		}
	})
}

func (n *UserNotification) sendAlertNotification(ctx context.Context, tasks []model.Task,
	orderInfo order.Order, userMap RecipientMap) error {
	var userIds []string
	for _, src := range tasks {
		receiver := userMap.GetID(src.UserID)
		if receiver == "" {
			continue
		}
		userIds = append(userIds, receiver)
	}

	params, err := structpb.NewStruct(orderInfo.NotificationConf.TemplateParams)
	if err != nil {
		return fmt.Errorf("解析模版数据失败: %w", err)
	}

	// 发送消息
	_, err = n.notificationSvc.SendNotification(ctx, &notificationv1.SendNotificationRequest{
		Notification: &notificationv1.Notification{
			Key:            "ecmdb",
			Receivers:      userIds,
			Channel:        order.ChannelToDomainProto(orderInfo.NotificationConf.Channel),
			TemplateParams: params,
			TemplateId:     orderInfo.NotificationConf.TemplateID,
		},
	})

	if err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	return nil
}
