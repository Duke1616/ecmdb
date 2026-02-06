package feishu

import (
	"context"
	"encoding/json"
	"fmt"

	notificationv1 "github.com/Duke1616/ecmdb/api/proto/gen/notification/v1"
	"github.com/Duke1616/ecmdb/internal/event/domain"
	"github.com/Duke1616/ecmdb/internal/event/service/provider"
	"github.com/Duke1616/ecmdb/internal/event/service/strategy"
	"github.com/Duke1616/enotify/notify/feishu/card"
	"github.com/google/uuid"
	"github.com/gotomicro/ego/core/elog"
	"google.golang.org/protobuf/types/known/structpb"
)

type grpcProvider struct {
	notification notificationv1.NotificationServiceClient
	logger       *elog.Component
}

func NewGRPCProvider(notification notificationv1.NotificationServiceClient) provider.Provider {
	return &grpcProvider{
		notification: notification,
		logger:       elog.DefaultLogger.With(elog.FieldComponentName("grpc_provider")),
	}
}

func (f *grpcProvider) Send(ctx context.Context, src domain.Notification) (bool, error) {
	builderMsg := card.NewApprovalCardBuilder().
		SetToTitle(src.Template.Title).
		SetToFields(toCardFields(src.Template.Fields)).
		SetToHideForm(src.Template.HideForm).
		SetInputFields(toCardInputFields(src.Template.InputFields)).
		SetToCallbackValue(toCardValues(src.Template.Values)).Build()
	
	var rawMap map[string]interface{}
	bytes, err := json.Marshal(builderMsg)
	if err != nil {
		return false, err
	}
	if err = json.Unmarshal(bytes, &rawMap); err != nil {
		return false, err
	}

	params, err := structpb.NewStruct(rawMap)
	if err != nil {
		return false, err
	}

	msg, err := f.notification.SendNotification(ctx, &notificationv1.SendNotificationRequest{Notification: &notificationv1.Notification{
		BizId:          notificationv1.Business_TICKET,
		Key:            uuid.New().String(),
		Receivers:      []string{src.Receiver},
		Channel:        notificationv1.Channel_FEISHU_CARD,
		TemplateId:     getTemplateID(src.Template.Name),
		TemplateParams: params,
	}})

	if err != nil || msg.Status != notificationv1.SendStatus_SUCCEEDED {
		return false, fmt.Errorf("消息发送失败: %v", msg)
	}

	return true, nil
}

func getTemplateID(templateName string) int64 {
	var templateID int64
	switch templateName {
	case strategy.LarkTemplateApprovalName:
		return templateID
	case strategy.LarkTemplateApprovalRevokeName:
		return templateID
	case strategy.LarkTemplateCC:
		return templateID
	default:
		return -1
	}
}
