package event

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
	"github.com/xen0n/go-workwx"
)

type WechatOrderConsumer struct {
	svc      service.Service
	consumer mq.Consumer
	logger   *elog.Component
}

func NewWechatOrderConsumer(svc service.Service, q mq.MQ) (*WechatOrderConsumer, error) {
	groupID := "wechat_order"
	consumer, err := q.Consumer(WechatOrderEventName, groupID)
	if err != nil {
		return nil, err
	}
	return &WechatOrderConsumer{
		svc:      svc,
		consumer: consumer,
		logger:   elog.DefaultLogger,
	}, nil
}

func (c *WechatOrderConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				elog.Error("同步事件失败", elog.Any("err", err))
			}
		}
	}()
}

func (c *WechatOrderConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	var evt workwx.OAApprovalDetail
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	return nil
}

func (c *WechatOrderConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}
