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

type NotifyParamsBuilder struct {
	params NotifyParams
}

// NewNotifyParamsBuilder 创建一个新 Builder
func NewNotifyParamsBuilder() *NotifyParamsBuilder {
	return &NotifyParamsBuilder{
		params: NotifyParams{
			Rules:      []rule.Rule{},
			Tasks:      []model.Task{},
			WantResult: make(map[string]interface{}),
		},
	}
}

// SetRules 设置规则
func (b *NotifyParamsBuilder) SetRules(rules []rule.Rule) *NotifyParamsBuilder {
	b.params.Rules = rules
	return b
}

// SetOrder 设置订单
func (b *NotifyParamsBuilder) SetOrder(o order.Order) *NotifyParamsBuilder {
	b.params.Order = o
	return b
}

// SetTasks 设置任务
func (b *NotifyParamsBuilder) SetTasks(tasks []model.Task) *NotifyParamsBuilder {
	b.params.Tasks = tasks
	return b
}

// SetWantResult 设置期望结果
func (b *NotifyParamsBuilder) SetWantResult(result map[string]interface{}) *NotifyParamsBuilder {
	b.params.WantResult = result
	return b
}

// Build 构建 NotifyParams
func (b *NotifyParamsBuilder) Build() NotifyParams {
	return b.params
}

type NotifierIntegration interface {
	Builder(title string, users []user.User, template string, params NotifyParams) []notify.NotifierWrap
}
