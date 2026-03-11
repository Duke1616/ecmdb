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
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/ecmdb/pkg/resolve"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"google.golang.org/protobuf/types/known/structpb"
)

type UserNotification struct {
	BaseStrategy
	sender          sender.NotificationSender
	departmentSvc   department.Service
	notificationSvc notificationv1.NotificationServiceClient
	assigneeService *resolve.Engine
}

func NewUserNotification(base BaseStrategy, departmentSvc department.Service,
	sender sender.NotificationSender, notificationSvc notificationv1.NotificationServiceClient,
	assigneeService *resolve.Engine) *UserNotification {

	return &UserNotification{
		BaseStrategy:    base,
		sender:          sender,
		departmentSvc:   departmentSvc,
		notificationSvc: notificationSvc,
		assigneeService: assigneeService,
	}
}

func (n *UserNotification) Send(ctx context.Context, info Info) (notification.NotificationResponse, error) {
	// 1. 获取当前节点信息（利用 BaseStrategy 封装好的逻辑）
	nodes, rawProps, err := n.GetNodeProperty(info, info.CurrentNode.NodeID)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), err.Error()), fmt.Errorf("%w: %v", errs.ErrNodeNotConfigured, err)
	}

	property, err := easyflow.ToNodeProperty[easyflow.UserProperty](easyflow.Node{Properties: rawProps})
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), err.Error()), fmt.Errorf("%w: %v", errs.ErrNodeNotConfigured, err)
	}

	// 2. 并行获取所需元数据（规则、自动化结果、发起人）
	data, err := n.FetchRequiredData(ctx, info, nodes)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeFetchDataFailed), err.Error()), fmt.Errorf("%w: %v", errs.ErrFetchData, err)
	}

	// 3. 利用解析引擎解析审批人，并同步回流程节点（由 easy-workflow 接管后续的任务下发）
	targets := n.EnrichTargets(info, property.NormalizeAssignees())

	users, err := n.assigneeService.Resolve(ctx, targets)
	if err != nil {
		n.Logger.Error("解析审批人规则失败", elog.FieldErr(err), elog.String("node", info.CurrentNode.NodeID))
		return notification.NewErrorResponse(string(errs.ErrorCodeResolveRuleFailed), err.Error()), err
	}

	// 更新 CurrentNode.UserIDs 协助引擎分发任务
	info.CurrentNode.UserIDs = slice.Map(users, func(idx int, u user.User) string {
		return u.Username
	})

	// 4. 异步处理消息发送（因需等待任务记录创建）
	// 创建独立的 context，但保留 trace 信息
	sendCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	go func() {
		defer cancel()
		defer func() {
			if r := recover(); r != nil {
				n.Logger.Error("异步发送通知发生panic", elog.Any("recover", r))
			}
		}()
		n.asyncSendNotification(sendCtx, info, property, data, users)
	}()

	return notification.NotificationResponse{}, nil
}

func (n *UserNotification) asyncSendNotification(ctx context.Context, info Info,
	property easyflow.UserProperty, data *NotificationData, users []user.User) {
	// 1. 获取任务信息（带重试）
	tasks, err := n.FetchTasksWithRetry(ctx, info)
	if err != nil {
		n.Logger.Error("获取任务信息失败", elog.FieldErr(err), elog.Int("instanceId", info.InstID))
		return
	}

	// 2. 准备通知元数据（模板、标题、抄送自动处理）
	template, title := n.prepareNotificationMetadata(ctx, property, data, tasks)

	// 3. 全局通知校验（确保抄送自动通过已执行后校验）
	if ok := n.IsGlobalNotify(info.Workflow); !ok {
		return
	}

	// 4. 组装接收者映射
	userMap := n.AnalyzeUsers(users)

	// 5. 特殊处理：告警转工单通知
	if info.Order.Provide.IsAlert() {
		if err = n.sendAlertNotification(ctx, tasks, info.Order, userMap); err != nil {
			n.Logger.Error("告警转工单消息发送失败", elog.FieldErr(err))
		}
		return
	}

	// 6. 准备通知展示字段
	fields := n.PrepareCommonFields(info, data)

	// 7. 构建并批量发送消息
	notifications := n.buildNotifications(info, property, tasks, userMap, template, title, fields)
	if _, err = n.sender.BatchSend(ctx, notifications); err != nil {
		n.Logger.Warn("发送消息失败", elog.FieldErr(err))
	}
}

func (n *UserNotification) prepareNotificationMetadata(ctx context.Context, property easyflow.UserProperty,
	data *NotificationData, tasks []model.Task) (template workflow.NotifyType, title string) {
	title = rule.GenerateTitle(data.StartUser.DisplayName, data.TName)
	template = LarkTemplateApprovalName
	return
}

func (n *UserNotification) buildNotifications(info Info, property easyflow.UserProperty,
	tasks []model.Task, userMap map[string]string,
	template workflow.NotifyType, title string, fields []notification.Field) []notification.Notification {
	return slice.Map(tasks, func(idx int, src model.Task) notification.Notification {
		receiver, _ := userMap[src.UserID]
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
	oInfo order.Order, userMap map[string]string) error {
	userIds := slice.Map(tasks, func(idx int, src model.Task) string {
		receiver, _ := userMap[src.UserID]
		return receiver
	})

	params, err := structpb.NewStruct(oInfo.NotificationConf.TemplateParams)
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
