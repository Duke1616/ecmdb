package method

import (
	"github.com/Bunny3th/easy-workflow/workflow/model"
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

const (
	// FeishuTemplateApprovalName 正常审批通知
	FeishuTemplateApprovalName = "feishu-card-callback"
	// FeishuTemplateCC 抄送通知
	FeishuTemplateCC = "feishu-card-cc"
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

	nc, err := feishu.NewCreateFeishuNotifyByClient(lark)
	if err != nil {
		return nil, err
	}

	return &FeishuUserNotify{
		tmpl:     tmpl,
		tmplName: "feishu-card-callback",
		Nc:       nc,
	}, nil
}

// TODO title 生成规则，不同情况下应有不同的样子，比如是发送给自己的要展示 你的XXX申请 抄送给你
func (n *FeishuUserNotify) generate(userId string, title string, fields []card.Field,
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

func (n *FeishuUserNotify) Builder(title string, users []user.User, template string, params NotifyParams) []notify.NotifierWrap {
	// 获取自定义字段
	fields := rule.GetFields(params.Rules, params.Order.Provide.ToUint8(), params.Order.Data)

	// 解析飞书用户信息
	userMap := n.analyzeUsers(users)

	// 生成发送消息的结构
	messages := slice.Map(params.Tasks, func(idx int, src model.Task) notify.NotifierWrap {
		uid, _ := userMap[src.UserID]
		cardVal := []card.Value{
			{
				Key:   "order_id",
				Value: params.Order.Id,
			},
			{
				Key:   "task_id",
				Value: src.TaskID,
			},
		}

		return n.generate(uid, title, fields, cardVal, template)
	})

	return messages
}

// analyzeUsers 解析用户，把 ID 转换为飞书 ID
func (n *FeishuUserNotify) analyzeUsers(users []user.User) map[string]string {
	return slice.ToMapV(users, func(element user.User) (string, string) {
		return element.Username, element.FeishuInfo.UserId
	})
}
