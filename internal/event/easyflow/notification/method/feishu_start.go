package method

import (
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/enotify/notify"
	"github.com/Duke1616/enotify/notify/feishu"
	"github.com/Duke1616/enotify/notify/feishu/card"
	feishuMsg "github.com/Duke1616/enotify/notify/feishu/message"
	"github.com/Duke1616/enotify/template"
	"github.com/ecodeclub/ekit/slice"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type FeishuStartNotify struct {
	Nc   notify.Notifier[*larkim.CreateMessageReq]
	tmpl *template.Template
}

func NewFeishuStartNotify(lark *lark.Client) (*FeishuStartNotify, error) {
	tmpl, err := template.FromGlobs([]string{})
	if err != nil {
		return nil, err
	}

	nc, err := feishu.NewCreateFeishuNotifyByClient(lark)
	if err != nil {
		return nil, err
	}

	return &FeishuStartNotify{
		tmpl: tmpl,
		Nc:   nc,
	}, nil
}

func (n *FeishuStartNotify) Builder(title string, users []user.User, template string, params NotifyParams) []notify.NotifierWrap {
	// 获取自定义字段
	fields := rule.GetFields(params.Rules, params.Order.Provide.ToUint8(), params.Order.Data)

	// 解析飞书用户信息
	userMap := n.analyzeUsers(users)

	// 生成发送消息的结构
	messages := slice.Map(users, func(idx int, src user.User) notify.NotifierWrap {
		uid, _ := userMap[src.Username]
		cardVal := []card.Value{
			{
				Key:   "order_id",
				Value: params.Order.Id,
			},
			{
				Key:   "task_id",
				Value: "100001",
			},
		}

		return n.generate(uid, title, fields, cardVal, template)
	})

	return messages
}

func (n *FeishuStartNotify) generate(userId string, title string, fields []card.Field,
	cardVal []card.Value, template string) notify.NotifierWrap {
	return notify.WrapNotifierDynamic(n.Nc, func() (notify.BasicNotificationMessage[*larkim.CreateMessageReq], error) {
		return feishuMsg.NewCreateFeishuMessage(
			"user_id", userId,
			feishu.NewFeishuCustomCard(n.tmpl, template,
				card.NewApprovalCardBuilder().
					SetToTitle(title).
					SetToFields(fields).
					SetToCallbackValue(cardVal).Build(),
			),
		), nil
	})
}

// analyzeUsers 解析用户，把 ID 转换为飞书 ID
func (n *FeishuStartNotify) analyzeUsers(users []user.User) map[string]string {
	return slice.ToMapV(users, func(element user.User) (string, string) {
		return element.Username, element.FeishuInfo.UserId
	})
}
