package method

import (
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/enotify/notify"
)

const (
	// FeishuTemplateApprovalName 正常审批通知
	FeishuTemplateApprovalName = "feishu-card-callback"
	// FeishuTemplateApprovalRevokeName 带有撤销的审批通知
	FeishuTemplateApprovalRevokeName = "feishu-card-revoke"
	// FeishuTemplateCC 抄送通知
	FeishuTemplateCC = "feishu-card-cc"
)

type NotifyParams struct {
	Rules      []rule.Rule
	Order      order.Order
	Tasks      []model.Task
	WantResult map[string]interface{}
}

type NotifierIntegration interface {
	Builder(title string, users []user.User, template string, params NotifyParams) []notify.NotifierWrap
}
