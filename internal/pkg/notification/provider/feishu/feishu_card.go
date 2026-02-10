package feishu

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/provider"
	"github.com/Duke1616/enotify/notify"
	"github.com/Duke1616/enotify/notify/feishu"
	"github.com/Duke1616/enotify/notify/feishu/card"
	"github.com/Duke1616/enotify/notify/feishu/message"
	"github.com/Duke1616/enotify/template"
	lark "github.com/larksuite/oapi-sdk-go/v3"
)

type larkCardProvider struct {
	handler notify.Handler
	tmpl    *template.Template
}

func NewLarkCardProvider(lark *lark.Client) (provider.Provider, error) {
	tmpl, err := template.FromGlobs([]string{})
	if err != nil {
		return nil, err
	}

	handler, err := feishu.NewHandler(lark)
	if err != nil {
		return nil, err
	}

	return &larkCardProvider{
		tmpl:    tmpl,
		handler: handler,
	}, nil
}

func (f *larkCardProvider) Send(ctx context.Context, src notification.Notification) (
	notification.NotificationResponse, error) {
	builder := f.buildBuilder(src)

	content := feishu.NewFeishuCustomCard(
		f.tmpl,
		string(src.Template.Name),
		builder.Build(),
	)

	msg := f.buildMessage(src, content)

	if err := f.handler.Send(ctx, msg); err != nil {
		return notification.NotificationResponse{}, err
	}

	return notification.NotificationResponse{}, nil
}

func (f *larkCardProvider) buildBuilder(src notification.Notification) card.Builder {
	builder := card.NewApprovalCardBuilder().
		SetToTitle(src.Template.Title).
		SetToFields(toCardFields(src.Template.Fields)).
		SetToHideForm(src.Template.HideForm).
		SetWantResult(src.Template.Remark).
		SetToCallbackValue(toCardValues(src.Template.Values)).
		SetInputFields(toCardInputFields(src.Template.InputFields))

	if src.IsProgressImageResult() {
		builder.SetImageKey(src.Template.ImageKey)
	}

	return builder
}

func (f *larkCardProvider) buildMessage(src notification.Notification, content message.Content) *notify.Message {
	if src.IsPatch() {
		return feishu.NewPatchBuilder(src.MessageID).
			SetReceiveIDType(feishu.ReceiveIDTypeUserID).
			SetContent(content).
			Build()
	}

	return feishu.NewCreateBuilder(src.Receiver).
		SetReceiveIDType(feishu.ReceiveIDTypeUserID).
		SetContent(content).
		Build()
}
