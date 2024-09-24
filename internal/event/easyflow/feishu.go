package easyflow

import (
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/enotify/notify"
	"github.com/Duke1616/enotify/notify/feishu"
	"github.com/Duke1616/enotify/notify/feishu/card"
	"github.com/Duke1616/enotify/template"
	"github.com/ecodeclub/ekit/slice"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type FeishuNotify struct {
	nc       notify.Notifier[*larkim.CreateMessageReq]
	tmpl     *template.Template
	tmplName string
}

func NewFeishuNotify(lark *lark.Client) (*FeishuNotify, error) {
	tmpl, err := template.FromGlobs([]string{})
	if err != nil {
		return nil, err
	}

	nc, err := feishu.NewFeishuNotifyByClient(lark)
	if err != nil {
		return nil, err
	}

	return &FeishuNotify{
		tmpl:     tmpl,
		tmplName: "feishu-card-callback",
		nc:       nc,
	}, nil

}

func (n *FeishuNotify) getFields(rules []Rule, order order.Order) ([]card.Field, string) {
	ruleMap := slice.ToMap(rules, func(element Rule) string {
		return element.Field
	})

	// 拼接消息体
	var fields []card.Field
	num := 1
	for field, value := range order.Data {
		title := field
		val, ok := ruleMap[field]
		if ok {
			title = val.Title
		}

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

	return fields, order.TemplateName
}

func (n *FeishuNotify) generate(userId string, title string, fields []card.Field, cardVal []card.Value) notify.NotifierWrap {
	return notify.WrapNotifierDynamic(n.nc, func() (notify.BasicNotificationMessage[*larkim.CreateMessageReq], error) {
		return feishu.NewFeishuMessage(
			"user_id", userId,
			feishu.NewFeishuCustomCard(n.tmpl, n.tmplName,
				card.NewApprovalCardBuilder().
					SetToTitle(title).
					SetToFields(fields).
					SetToCallbackValue(cardVal).Build(),
			),
		), nil
	})
}

func (n *FeishuNotify) builder(rules []Rule, order order.Order, users []user.User, tasks []model.Task) []notify.NotifierWrap {
	// 获取自定义字段
	fields, title := n.getFields(rules, order)

	// 解析飞书用户信息
	userMap := n.analyzeUsers(users)

	// 生成发送消息的结构
	messages := slice.Map(tasks, func(idx int, src model.Task) notify.NotifierWrap {
		uid, _ := userMap[src.UserID]
		cardVal := []card.Value{
			{
				Key:   "task_id",
				Value: src.TaskID,
			},
		}

		return n.generate(uid, title, fields, cardVal)
	})

	return messages
}

// analyzeUsers 解析用户，把 ID 转换为飞书 ID
func (n *FeishuNotify) analyzeUsers(users []user.User) map[string]string {
	return slice.ToMapV(users, func(element user.User) (string, string) {
		return element.Username, element.FeishuInfo.UserId
	})
}
