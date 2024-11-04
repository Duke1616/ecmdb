package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order/internal/event"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	templateSvc "github.com/Duke1616/ecmdb/internal/template"

	"github.com/Duke1616/enotify/notify"
	"github.com/Duke1616/enotify/notify/feishu"
	"github.com/Duke1616/enotify/notify/feishu/card"
	feishuMsg "github.com/Duke1616/enotify/notify/feishu/message"
	"github.com/Duke1616/enotify/template"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"

	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
	"github.com/larksuite/oapi-sdk-go/v3"
	"strconv"
)

type FeishuCallbackEventConsumer struct {
	Nc  notify.Notifier[*larkim.PatchMessageReq]
	Svc service.Service

	tmpl        *template.Template
	tmplName    string
	engineSvc   engineSvc.Service
	templateSvc templateSvc.Service
	consumer    mq.Consumer
	lark        *lark.Client
	logger      *elog.Component
}

func NewFeishuCallbackEventConsumer(q mq.MQ, engineSvc engineSvc.Service, service service.Service,
	templateSvc templateSvc.Service, lark *lark.Client) (*FeishuCallbackEventConsumer, error) {
	groupID := "feishu_callback"
	consumer, err := q.Consumer(event.FeishuCallbackEventName, groupID)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.FromGlobs([]string{})
	if err != nil {
		return nil, err
	}

	nc, err := feishu.NewPatchFeishuNotifyByClient(lark)
	if err != nil {
		return nil, err
	}

	return &FeishuCallbackEventConsumer{
		consumer:    consumer,
		engineSvc:   engineSvc,
		Nc:          nc,
		templateSvc: templateSvc,
		Svc:         service,
		tmpl:        tmpl,
		tmplName:    "feishu-card-want",
		lark:        lark,
		logger:      elog.DefaultLogger,
	}, nil
}

func (c *FeishuCallbackEventConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				c.logger.Error("同步飞书回调事件失败", elog.Any("err", err))
			}
		}
	}()
}

func (c *FeishuCallbackEventConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}
	var evt event.FeishuCallback
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	taskId, err := strconv.Atoi(evt.TaskId)
	if err != nil {
		return err
	}

	// 处理消息
	if evt.Comment == "" {
		evt.Comment = "无"
	}

	wantResult := fmt.Sprintf("你已同意该申请, 批注：%s", evt.Comment)
	switch evt.Action {
	case "pass":
		err = c.engineSvc.Pass(ctx, taskId, evt.Comment)
		if err != nil {
			wantResult = "你的节点任务已经结束，无法进行审批，详情登录 ECMDB 平台查看"
			c.logger.Error("飞书回调消息，同意工单失败", elog.FieldErr(err))
		}
	case "reject":
		err = c.engineSvc.Reject(ctx, taskId, evt.Comment)
		if err != nil {
			wantResult = "你的节点任务已经结束，无法进行审批，详情登录 ECMDB 平台查看"
			c.logger.Error("飞书回调消息，驳回工单失败", elog.FieldErr(err))
		}
	}

	return c.withdraw(ctx, evt.MessageId, evt.OrderId, wantResult)
}

func (c *FeishuCallbackEventConsumer) withdraw(ctx context.Context, messageId string, orderId string, wantResult string) error {
	// 获取模版详情信息
	orderIdInt, _ := strconv.ParseInt(orderId, 10, 64)

	fOrder, err := c.Svc.Detail(ctx, orderIdInt)
	if err != nil {
		return err
	}

	t, err := c.templateSvc.DetailTemplate(ctx, fOrder.TemplateId)
	if err != nil {
		return err
	}

	rules, err := rule.ParseRules(t.Rules)
	if err != nil {
		return err
	}
	fields := rule.GetFields(rules, fOrder.Provide.ToUint8(), fOrder.Data)

	notifyWrap := notify.WrapNotifierDynamic(c.Nc, func() (notify.BasicNotificationMessage[*larkim.PatchMessageReq], error) {
		return feishuMsg.NewPatchFeishuMessage(
			messageId,
			feishu.NewFeishuCustomCard(c.tmpl, c.tmplName,
				card.NewApprovalCardBuilder().
					SetToFields(fields).
					SetWantResult(wantResult).
					Build(),
			),
		), nil
	})

	ok, err := notifyWrap.Send(ctx)
	if !ok {
		c.logger.Error("修改飞书消息失败")
	}
	return err
}

func (c *FeishuCallbackEventConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}
