package feishu

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/event/domain"
	"github.com/Duke1616/ecmdb/internal/event/service/provider"
	"github.com/Duke1616/enotify/notify"
	"github.com/Duke1616/enotify/notify/feishu"
	"github.com/Duke1616/enotify/notify/feishu/card"
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

func (f *larkCardProvider) Send(ctx context.Context, src domain.Notification) (domain.NotificationResponse, error) {
	msg := feishu.NewCreateBuilder(src.Receiver).SetReceiveIDType(feishu.ReceiveIDTypeUserID).
		SetContent(feishu.NewFeishuCustomCard(f.tmpl, src.Template.Name,
			card.NewApprovalCardBuilder().
				SetToTitle(src.Template.Title).
				SetToFields(toCardFields(src.Template.Fields)).
				SetToHideForm(src.Template.HideForm).
				SetInputFields(toCardInputFields(src.Template.InputFields)).
				SetToCallbackValue(toCardValues(src.Template.Values)).
				Build(),
		)).
		Build()

	if err := f.handler.Send(ctx, msg); err != nil {
		return domain.NotificationResponse{}, err
	}

	return domain.NotificationResponse{}, nil
}
