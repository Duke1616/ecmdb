package event

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
	"github.com/Duke1616/ecmdb/internal/template/internal/service"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
	"github.com/xen0n/go-workwx"
)

type WechatApprovalCallbackConsumer struct {
	svc      service.Service
	consumer mq.Consumer
	producer WechatOrderEventProducer
	workApp  *workwx.WorkwxApp
	logger   *elog.Component
}

func NewWechatApprovalCallbackConsumer(svc service.Service, q mq.MQ, p WechatOrderEventProducer, workApp *workwx.WorkwxApp) (*WechatApprovalCallbackConsumer, error) {
	groupID := "wechat_oa_callback"
	consumer, err := q.Consumer(WechatCallbackEventName, groupID)
	if err != nil {
		return nil, err
	}
	return &WechatApprovalCallbackConsumer{
		svc:      svc,
		consumer: consumer,
		logger:   elog.DefaultLogger,
		producer: p,
		workApp:  workApp,
	}, nil
}

func (c *WechatApprovalCallbackConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				elog.Error("创建企业微信工单事件失败", elog.Any("err", err))
			}
		}
	}()
}

func (c *WechatApprovalCallbackConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	var evt workwx.OAApprovalInfo
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	if _, err = c.svc.FindOrCreateByWechat(ctx, domain.WechatInfo{
		TemplateId:   evt.TemplateID,
		TemplateName: evt.SpName,
		SpNo:         evt.SpNo,
	}); err != nil {
		elog.Error("模版已经存在或新增模版失败", elog.Any("err", err))
		return err
	}

	return c.sendCreateOrderEvent(ctx, evt.SpNo)
}

func (c *WechatApprovalCallbackConsumer) sendCreateOrderEvent(ctx context.Context, spNo string) error {
	spInfo, err := c.workApp.GetOAApprovalDetail(spNo)
	if err != nil {
		return err
	}

	err = c.producer.Produce(ctx, spInfo)
	if err != nil {
		elog.Error("传输企业微信工单失败",
			elog.FieldErr(err),
			elog.Any("event", spInfo.SpNo),
		)
	}

	return err
}

func (c *WechatApprovalCallbackConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}
