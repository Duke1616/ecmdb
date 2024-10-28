package method

import (
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/wechat"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/enotify/notify"
	"github.com/Duke1616/enotify/notify/feishu"
	"github.com/Duke1616/enotify/notify/feishu/card"
	"github.com/Duke1616/enotify/template"
	"github.com/ecodeclub/ekit/slice"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/xen0n/go-workwx"
)

type FeishuUserNotify struct {
	Nc       notify.Notifier[*larkim.CreateMessageReq]
	tmpl     *template.Template
	tmplName string
}

func NewFeishuUserNotify(lark *lark.Client) (*FeishuUserNotify, error) {
	tmpl, err := template.FromGlobs([]string{})
	if err != nil {
		return nil, err
	}

	nc, err := feishu.NewFeishuNotifyByClient(lark)
	if err != nil {
		return nil, err
	}

	return &FeishuUserNotify{
		tmpl:     tmpl,
		tmplName: "feishu-card-callback",
		Nc:       nc,
	}, nil

}

func (n *FeishuUserNotify) getFields(rules []Rule, nOrder order.Order) []card.Field {
	ruleMap := slice.ToMap(rules, func(element Rule) string {
		return element.Field
	})

	// 拼接消息体
	num := 1
	var fields []card.Field

	// 判断不同平台的消息来源，进行处理
	switch nOrder.Provide {
	case order.SystemProvide:
		for field, value := range nOrder.Data {
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
	case order.WechatProvide:
		oaData, err := wechat.Unmarshal(nOrder.Data)
		if err != nil {
			return nil
		}

		for _, contents := range oaData.ApplyData.Contents {
			key := contents.Title[0].Text

			switch contents.Control {
			case "Selector":
				switch contents.Value.Selector.Type {
				case "single":
					fields = append(fields, card.Field{
						IsShort: true,
						Tag:     "lark_md",
						Content: fmt.Sprintf(`**%s:**\n%v`, key, contents.Value.Selector.Options[0].Value[0].Text),
					})
				case "multi":
					value := slice.Map(contents.Value.Selector.Options, func(idx int,
						src workwx.OAContentSelectorOption) string {
						return src.Value[0].Text
					})

					fields = append(fields, card.Field{
						IsShort: true,
						Tag:     "lark_md",
						Content: fmt.Sprintf(`**%s:**\n%v`, key, value),
					})
				}
			case "Textarea":
				fields = append(fields, card.Field{
					IsShort: true,
					Tag:     "lark_md",
					Content: fmt.Sprintf(`**%s:**\n%v`, key, contents.Value.Text),
				})
			case "default":
				fmt.Println("不符合筛选规则")
			}

			if num%2 == 0 {
				fields = append(fields, card.Field{
					IsShort: false,
					Tag:     "lark_md",
					Content: "",
				})
			}

			num++
		}
	}

	return fields
}

// TODO title 生成规则，不同情况下应有不同的样子，比如是发送给自己的要展示 你的XXX申请 抄送给你
func (n *FeishuUserNotify) generate(userId string, title string, fields []card.Field,
	cardVal []card.Value) notify.NotifierWrap {
	return notify.WrapNotifierDynamic(n.Nc, func() (notify.BasicNotificationMessage[*larkim.CreateMessageReq], error) {
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

func (n *FeishuUserNotify) Builder(title string, users []user.User, params NotifyParams) []notify.NotifierWrap {
	// 获取自定义字段
	fields := n.getFields(params.Rules, params.Order)

	// 解析飞书用户信息
	userMap := n.analyzeUsers(users)

	// 生成发送消息的结构
	messages := slice.Map(params.Tasks, func(idx int, src model.Task) notify.NotifierWrap {
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
func (n *FeishuUserNotify) analyzeUsers(users []user.User) map[string]string {
	return slice.ToMapV(users, func(element user.User) (string, string) {
		return element.Username, element.FeishuInfo.UserId
	})
}
