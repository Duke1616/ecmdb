package node

import (
	"context"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/event/easyflow/notification/method"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/enotify/notify"
	"github.com/gotomicro/ego/core/elog"
)

type StartNotification struct {
	integrations []method.NotifyIntegration
	userSvc      user.Service
	templateSvc  template.Service

	logger *elog.Component
}

func NewStartNotification(userSvc user.Service, templateSvc template.Service, integrations []method.NotifyIntegration) (*StartNotification, error) {
	return &StartNotification{
		integrations: integrations,
		userSvc:      userSvc,
		templateSvc:  templateSvc,
		logger:       elog.DefaultLogger,
	}, nil
}

func (s *StartNotification) Send(ctx context.Context, nOrder order.Order, wf workflow.Workflow,
	instanceId int, currentNode *model.Node) (bool, error) {
	// 获取当前节点信息
	property, err := getNodeProperty[easyflow.StartProperty](wf, currentNode.NodeID)
	if err != nil {
		return false, err
	}

	// 判断开始节点是否需要发送消息通知
	if ok := s.isNotify(property, instanceId); !ok {
		return false, nil
	}

	// 解析配置
	rules, err := s.getRules(ctx, nOrder)
	if err != nil {
		return false, err
	}

	// 获取工单创建用户
	startUser, err := s.userSvc.FindByUsername(ctx, nOrder.CreateBy)
	if err != nil {
		return false, err
	}

	var messages []notify.NotifierWrap
	title := rule.GenerateTitle("你提交的", nOrder.TemplateName)
	for _, integration := range s.integrations {
		if integration.Name == "feishu_start" {
			messages = integration.Notifier.Builder(title, []user.User{startUser},
				method.FeishuTemplateApprovalRevokeName, method.NewNotifyParamsBuilder().
					SetRules(rules).
					SetOrder(nOrder).
					Build())
			break
		}
	}

	if ok, er := send(context.Background(), messages); er != nil || !ok {
		s.logger.Warn("发送消息失败",
			elog.Any("error", er),
		)

		return false, nil
	}

	return true, nil
}

func (s *StartNotification) isNotify(sp easyflow.StartProperty, instanceId int) bool {
	if !sp.IsNotify {
		s.logger.Warn("流程控制【开始节点】未开启消息通知能力",
			elog.Any("instId", instanceId),
		)
		return false
	}

	return true
}

// isNotify 获取模版的字段信息
func (s *StartNotification) getRules(ctx context.Context, order order.Order) ([]rule.Rule, error) {
	// 获取模版详情信息
	t, err := s.templateSvc.DetailTemplate(ctx, order.TemplateId)
	if err != nil {
		return nil, err
	}

	rules, err := rule.ParseRules(t.Rules)

	if err != nil {
		return nil, err
	}

	return rules, nil
}
