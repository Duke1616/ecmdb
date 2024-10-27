package method

import lark "github.com/larksuite/oapi-sdk-go/v3"

type NotifyIntegration struct {
	Notifier NotifierIntegration
	Name     string
}

func NewNotifyIntegration(n NotifierIntegration, name string) NotifyIntegration {
	return NotifyIntegration{
		Notifier: n,
		Name:     name,
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
