package feishu

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/provider"
	"github.com/Duke1616/enotify/notify"
	"github.com/Duke1616/enotify/notify/feishu"
	lark "github.com/larksuite/oapi-sdk-go/v3"
)

type larkTextProvider struct {
	handler notify.Handler
}

func NewLarkTextProvider(lark *lark.Client) (provider.Provider, error) {
	handler, err := feishu.NewHandler(lark)
	if err != nil {
		return nil, err
	}

	return &larkTextProvider{
		handler: handler,
	}, nil
}

func (f *larkTextProvider) Send(ctx context.Context, src notification.Notification) (
	notification.NotificationResponse, error) {
	content := fmt.Sprintf(`{"text": "%s"}`, src.Template.Text)
	msg := feishu.NewCreateBuilder(src.Receiver).SetReceiveIDType(feishu.ReceiveIDTypeUserID).
		SetContent(feishu.NewFeishuCustom("text", content)).Build()

	if err := f.handler.Send(ctx, msg); err != nil {
		return notification.NotificationResponse{}, fmt.Errorf("触发发送信息失败: %w", err)
	}

	return notification.NotificationResponse{}, nil
}
