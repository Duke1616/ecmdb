package easyflow

import (
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/enotify/notify"
	lark "github.com/larksuite/oapi-sdk-go/v3"
)

type NotifierIntegration interface {
	builder(rules []Rule, order order.Order, startUser string, users []user.User, tasks []model.Task) []notify.NotifierWrap
}

type NotifyIntegration struct {
	notifier NotifierIntegration
	name     string
}

func NewNotifyIntegration(n NotifierIntegration, name string) NotifyIntegration {
	return NotifyIntegration{
		notifier: n,
		name:     name,
	}
}

// BuildReceiverIntegrations 整合消息通知渠道
func BuildReceiverIntegrations(larkC *lark.Client) ([]NotifyIntegration, error) {
	var (
		integrations []NotifyIntegration
		add          = func(name string, f func() (NotifierIntegration, error)) {
			n, err := f()
			if err != nil {
				return
			}
			integrations = append(integrations, NewNotifyIntegration(n, name))
		}
	)

	add("feishu", func() (NotifierIntegration, error) { return NewFeishuNotify(larkC) })
	return integrations, nil
}
