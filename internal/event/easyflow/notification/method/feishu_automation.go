package method

import (
	"fmt"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/enotify/notify"
	"github.com/Duke1616/enotify/notify/feishu"
	"github.com/Duke1616/enotify/notify/feishu/card"
	"github.com/Duke1616/enotify/notify/feishu/message"
	"github.com/Duke1616/enotify/template"
	"github.com/ecodeclub/ekit/slice"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type FeishuAutomationNotify struct {
	Nc   notify.Notifier[*larkim.CreateMessageReq]
	tmpl *template.Template
}

func NewFeishuAutomationNotify(lark *lark.Client) (*FeishuAutomationNotify, error) {
	tmpl, err := template.FromGlobs([]string{})
	if err != nil {
		return nil, err
	}

	nc, err := feishu.NewCreateFeishuNotifyByClient(lark)
	if err != nil {
		return nil, err
	}

	return &FeishuAutomationNotify{
		tmpl: tmpl,
		Nc:   nc,
	}, nil

}

func (n *FeishuAutomationNotify) getFields(wantResult map[string]interface{}) []card.Field {
	num := 1
	var fields []card.Field

	for field, value := range wantResult {
		title := field

		fields = append(fields, card.Field{
			IsShort: true,
			Tag:     "lark_md",
			Content: fmt.Sprintf(`**%s:**\n%v`, title, value),
		})

		if num%2 == 0 {
			fields = append(fields, card.Field{
				IsShort: false,
				Tag:     "lark_md",
				Content: "",
			})
		}

		num++
	}

	return fields
}

// TODO title 生成规则，不同情况下应有不同的样子，比如是发送给自己的要展示 你的XXX申请 抄送给你
func (n *FeishuAutomationNotify) generate(userId string, title string, fields []card.Field,
	cardVal []card.Value, template string) notify.NotifierWrap {
	return notify.WrapNotifierDynamic(n.Nc, func() (notify.BasicNotificationMessage[*larkim.CreateMessageReq], error) {
		return message.NewCreateFeishuMessage(
			"user_id", userId,
			feishu.NewFeishuCustomCard(n.tmpl, template,
				card.NewApprovalCardBuilder().
					SetToTitle(title).
					SetToFields(fields).
					SetToHideForm(true).
					SetToCallbackValue(cardVal).Build(),
			),
		), nil
	})
}

func (n *FeishuAutomationNotify) Builder(title string, users []user.User, template string, params NotifyParams) []notify.NotifierWrap {
	// 获取自定义字段
	fields := n.getFields(params.WantResult)

	// 解析飞书用户信息
	userMap := n.analyzeUsers(users)

	// 生成发送消息的结构
	messages := slice.Map(users, func(idx int, src user.User) notify.NotifierWrap {
		uid, _ := userMap[src.Username]
		var cardVal []card.Value

		return n.generate(uid, title, fields, cardVal, template)
	})

	return messages
}

// analyzeUsers 解析用户，把 ID 转换为飞书 ID
func (n *FeishuAutomationNotify) analyzeUsers(users []user.User) map[string]string {
	return slice.ToMapV(users, func(element user.User) (string, string) {
		return element.Username, element.FeishuInfo.UserId
	})
}
