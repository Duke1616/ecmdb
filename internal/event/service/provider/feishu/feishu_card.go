package feishu

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/event/domain"
	"github.com/Duke1616/ecmdb/internal/event/service/provider"
	"github.com/Duke1616/enotify/notify"
	"github.com/Duke1616/enotify/notify/feishu"
	"github.com/Duke1616/enotify/notify/feishu/card"
	"github.com/Duke1616/enotify/notify/feishu/message"
	"github.com/Duke1616/enotify/template"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type feishuCardProvider struct {
	notify notify.Notifier[*larkim.CreateMessageReq]
	tmpl   *template.Template
}

func NewFeishuCardProvider(lark *lark.Client) (provider.Provider, error) {
	tmpl, err := template.FromGlobs([]string{})
	if err != nil {
		return nil, err
	}

	nc, err := feishu.NewCreateFeishuNotifyByClient(lark)
	if err != nil {
		return nil, err
	}

	return &feishuCardProvider{
		tmpl:   tmpl,
		notify: nc,
	}, nil
}

func (f *feishuCardProvider) Send(ctx context.Context, notification domain.Notification) (bool, error) {
	return notify.WrapNotifierDynamic(f.notify, func() (notify.BasicNotificationMessage[*larkim.CreateMessageReq], error) {
		return message.NewCreateFeishuMessage(
			"user_id", notification.Receiver,
			feishu.NewFeishuCustomCard(f.tmpl, notification.Template.Name,
				card.NewApprovalCardBuilder().
					SetToTitle(notification.Template.Title).
					SetToFields(notification.Template.Fields).
					SetToHideForm(notification.Template.HideForm).
					SetToCallbackValue(notification.Template.Values).Build(),
			),
		), nil
	}).Send(ctx)
}
