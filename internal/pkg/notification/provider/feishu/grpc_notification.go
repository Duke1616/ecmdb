package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	notificationv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/notification/v1"
	"github.com/Duke1616/ecmdb/internal/event/errs"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/provider"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/enotify/notify/feishu/card"
	"github.com/google/uuid"
	"github.com/gotomicro/ego/core/elog"
	"google.golang.org/protobuf/types/known/structpb"
)

type grpcProvider struct {
	notification notificationv1.NotificationServiceClient
	workflowSvc  workflow.Service
	logger       *elog.Component
}

func NewGRPCProvider(notification notificationv1.NotificationServiceClient, workflowSvc workflow.Service) provider.Provider {
	return &grpcProvider{
		notification: notification,
		workflowSvc:  workflowSvc,
		logger:       elog.DefaultLogger.With(elog.FieldComponentName("grpc_provider")),
	}
}

func (f *grpcProvider) Send(ctx context.Context, src notification.Notification) (notification.NotificationResponse, error) {
	// 如果是修改情况，直接退出
	if src.IsPatch() {
		return notification.NotificationResponse{}, fmt.Errorf("grpc notification 不支持修改模式")
	}

	// 1. 创建 Builder 并应用通用属性
	builder := card.NewApprovalCardBuilder().
		SetToTitle(src.Template.Title).
		SetToFields(toCardFields(src.Template.Fields)).
		SetToHideForm(src.Template.HideForm).
		SetWantResult(src.Template.Remark).
		SetToCallbackValue(toCardValues(src.Template.Values)).
		SetInputFields(toCardInputFields(src.Template.InputFields))

	// 2. 根据类型应用特殊属性
	if src.IsProgressImageResult() {
		builder.SetImageKey(src.Template.ImageKey)
	}

	var rawMap map[string]interface{}
	bytes, err := json.Marshal(builder.Build())
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeBuildFailed), err.Error()), fmt.Errorf("%w: %v", errs.ErrBuildMessage, err)
	}
	if err = json.Unmarshal(bytes, &rawMap); err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeParseFailed), err.Error()), fmt.Errorf("%w: %v", errs.ErrParseMessage, err)
	}

	params, err := structpb.NewStruct(rawMap)
	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeParseFailed), err.Error()), fmt.Errorf("%w: %v", errs.ErrParseMessage, err)
	}

	// 这里的超时时间设置非常短 (3s)，目的是快速探测 gRPC 服务状态。
	// 如果 gRPC 服务未启动或网络不通，应该尽快超时报错，从而触发上层 channel 的 Fallback 机制 (降级到 Feishu Card 直连)。
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 获取绑定的消息通知模版ID
	templateID, err := f.getTemplateID(ctx, src.WorkFlowID, src.Template.Name, src.Channel.String())
	if err != nil {
		return notification.NotificationResponse{}, err
	}
	msg, err := f.notification.SendNotification(ctx, &notificationv1.SendNotificationRequest{Notification: &notificationv1.Notification{
		BizId:          notificationv1.Business_TICKET,
		Key:            uuid.New().String(),
		Receivers:      []string{src.Receiver},
		Channel:        notificationv1.Channel_FEISHU_CARD,
		TemplateId:     templateID,
		TemplateParams: params,
	}})

	if err != nil {
		return notification.NewErrorResponse(string(errs.ErrorCodeServiceUnavailable), err.Error()), fmt.Errorf("%w: %v", errs.ErrNotificationUnavailable, err)
	}

	if msg.Status != notificationv1.SendStatus_SUCCEEDED {
		return notification.NewErrorResponseWithID(
			int64(msg.NotificationId),
			msg.Status.String(),
			string(errs.ErrorCodeUnknown),
			"消息发送未成功",
		), fmt.Errorf("%w: 状态=%v", errs.ErrNotificationFailed, msg.Status)
	}

	return notification.NewSuccessResponse(int64(msg.NotificationId), msg.Status.String()), nil
}

func (f *grpcProvider) getTemplateID(ctx context.Context, workflowID int64, templateName workflow.NotifyType,
	channel string) (int64, error) {
	effective, err := f.workflowSvc.AdminNotifyBinding().GetEffective(ctx, workflowID, templateName, channel)

	return effective.TemplateId, err
}
