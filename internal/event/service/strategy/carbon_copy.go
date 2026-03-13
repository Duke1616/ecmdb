package strategy

import (
	"context"
	"fmt"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/event/errs"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/sender"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/ecmdb/pkg/resolve"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

type CarbonCopyNotification struct {
	Service
	sender          sender.NotificationSender
	assigneeService *resolve.Engine
}

func NewCarbonCopyNotification(base Service, sender sender.NotificationSender,
	assigneeService *resolve.Engine) *CarbonCopyNotification {
	return &CarbonCopyNotification{
		Service:         base,
		sender:          sender,
		assigneeService: assigneeService,
	}
}

func (n *CarbonCopyNotification) Send(ctx context.Context, info Info) (notification.NotificationResponse, error) {
	// 1. 获取节点属性
	nodes, rawProps, err := n.GetNodeProperty(info, info.CurrentNode.NodeID)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), err.Error()), err
	}
	property, err := easyflow.ToNodeProperty[easyflow.UserProperty](easyflow.Node{Properties: rawProps})
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeNodeNotConfigured), fmt.Sprintf("节点属性解析失败: %v", err)), err
	}

	// 2. 加载基础数据
	data, err := n.FetchRequiredData(ctx, info, nodes)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeFetchDataFailed), err.Error()), err
	}

	// 3. 解析抄送人员并自动同步到节点
	users, err := n.ResolveAssignees(ctx, &info, property.NormalizeAssignees())
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeResolveRuleFailed), err.Error()), err
	}

	// 4. 异步处理
	n.SafeGo(ctx, 3*time.Minute, func(sendCtx context.Context) {
		n.asyncHandleCarbonCopy(sendCtx, info, data, NewRecipientMap(users, info.Channel))
	})

	return notification.NewSuccessResponse(0, "success"), nil
}

func (n *CarbonCopyNotification) asyncHandleCarbonCopy(ctx context.Context, info Info, data *NotificationData, userMap RecipientMap) {
	// 1. 获取任务
	tasks, err := n.FetchTasksWithRetry(ctx, info)
	if err != nil {
		n.Logger().Error("CarbonCopy 获取任务失败", elog.FieldErr(err))
		return
	}

	// 2. 发送消息
	if n.IsGlobalNotify(info.Workflow) {
		title := rule.GenerateCCTitle(data.StartUser.DisplayName, data.TName)
		fields := n.PrepareCommonFields(info, data)

		ns := slice.Map(tasks, func(idx int, src model.Task) notification.Notification {
			receiver := userMap.GetID(src.UserID)
			return notification.Notification{
				Channel:    info.Channel,
				WorkFlowID: info.Workflow.Id,
				Receiver:   receiver,
				Template: notification.Template{
					Name:     LarkTemplateCC,
					Title:    title,
					Fields:   fields,
					Values:   []notification.Value{{Key: "order_id", Value: info.Order.Id}},
					HideForm: true,
				},
			}
		})

		for _, msg := range ns {
			if _, err = n.sender.Send(ctx, msg); err != nil {
				n.Logger().Error("CarbonCopyNotification 消息发送失败", elog.FieldErr(err), elog.String("receiver", msg.Receiver))
			}
		}
	}

	// 3. 立即自动通过
	for _, t := range tasks {
		if err = n.PassTask(ctx, t.TaskID, "抄送节点自动通过"); err != nil {
			n.Logger().Error("抄送节点自动通过失败", elog.FieldErr(err), elog.Any("taskId", t.TaskID))
		}

		if t.IsCosigned != 1 {
			return
		}
	}
}
