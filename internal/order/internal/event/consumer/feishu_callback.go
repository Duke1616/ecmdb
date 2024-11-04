package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order/internal/event"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
	"github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"strconv"
)

type FeishuCallbackEventConsumer struct {
	engineSvc engineSvc.Service
	consumer  mq.Consumer
	lark      *lark.Client
	logger    *elog.Component
}

func NewFeishuCallbackEventConsumer(q mq.MQ, engineSvc engineSvc.Service, lark *lark.Client) (*FeishuCallbackEventConsumer, error) {
	groupID := "feishu_callback"
	consumer, err := q.Consumer(event.FeishuCallbackEventName, groupID)
	if err != nil {
		return nil, err
	}

	return &FeishuCallbackEventConsumer{
		consumer:  consumer,
		engineSvc: engineSvc,
		lark:      lark,
		logger:    elog.DefaultLogger,
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
	switch evt.Action {
	case "pass":
		err = c.engineSvc.Pass(ctx, taskId, evt.Comment)
		if err != nil {
			c.logger.Error("飞书回调消息，同意工单失败", elog.FieldErr(err))
		}
	case "reject":
		err = c.engineSvc.Reject(ctx, taskId, evt.Comment)
		if err != nil {
			c.logger.Error("飞书回调消息，驳回工单失败", elog.FieldErr(err))
		}
	}

	return c.withdraw(ctx, evt.MessageId)
}

func (c *FeishuCallbackEventConsumer) withdraw(ctx context.Context, messageId string) error {
	req := larkim.NewDeleteMessageReqBuilder().
		MessageId(messageId).
		Build()

	resp, err := c.lark.Im.Message.Delete(ctx, req)
	if err != nil {
		return err
	}

	// 服务端错误处理
	if !resp.Success() {
		return fmt.Errorf("撤回消息服务端错误，code: %d, msg: %s, requestId: %s",
			resp.Code, resp.Msg, resp.RequestId())
	}

	return nil
}

func (c *FeishuCallbackEventConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}
