package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/event"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
	"github.com/xen0n/go-workwx"
)

type WechatOrderConsumer struct {
	svc         service.Service
	templateSvc template.Service
	consumer    mq.Consumer
	logger      *elog.Component
}

func NewWechatOrderConsumer(svc service.Service, templateSvc template.Service, q mq.MQ) (*WechatOrderConsumer, error) {
	groupID := "wechat_create_order"
	consumer, err := q.Consumer(event.WechatOrderEventName, groupID)
	if err != nil {
		return nil, err
	}
	return &WechatOrderConsumer{
		svc:         svc,
		consumer:    consumer,
		templateSvc: templateSvc,
		logger:      elog.DefaultLogger,
	}, nil
}

func (c *WechatOrderConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				elog.Error("同步企业微信工单创建工单事件失败", elog.Any("err", err))
			}
		}
	}()
}

func (c *WechatOrderConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	// 接收数据
	var evt workwx.OAApprovalDetail
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	// 转换成 map[string]interface{}
	data, err := convert(evt)
	if err != nil {
		return fmt.Errorf("数据转换失败: %w", err)
	}

	// 查看模版信息
	t, err := c.templateSvc.DetailTemplateByExternalTemplateId(ctx, evt.TemplateID)
	if err != nil {
		return fmt.Errorf("查看模版信息错误: %w", err)
	}

	// 创建工单
	err = c.svc.CreateOrder(ctx, domain.Order{
		CreateBy:     "企业微信",
		TemplateName: t.Name,
		TemplateId:   t.Id,
		WorkflowId:   t.WorkflowId,
		Data:         data,
		Status:       domain.START,
		Provide:      domain.WECHAT,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *WechatOrderConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}

func convert(evt workwx.OAApprovalDetail) (map[string]interface{}, error) {
	evtData, err := json.Marshal(evt)
	if err != nil {
		fmt.Println("Error marshalling struct:", err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(evtData, &data)
	return data, err
}
