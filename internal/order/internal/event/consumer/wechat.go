package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/event"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/internal/pkg/wechat"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
	"github.com/xen0n/go-workwx"
	"go.mongodb.org/mongo-driver/mongo"
)

type WechatOrderConsumer struct {
	svc         service.Service
	templateSvc template.Service
	userSvc     user.Service
	consumer    mq.Consumer
	logger      *elog.Component
}

func NewWechatOrderConsumer(svc service.Service, templateSvc template.Service, userSvc user.Service, q mq.MQ) (*WechatOrderConsumer, error) {
	groupID := "wechat_create_order"
	consumer, err := q.Consumer(event.WechatOrderEventName, groupID)
	if err != nil {
		return nil, err
	}
	return &WechatOrderConsumer{
		svc:         svc,
		consumer:    consumer,
		userSvc:     userSvc,
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
	data, err := wechat.Marshal(evt)
	if err != nil {
		return fmt.Errorf("数据转换失败: %w", err)
	}

	// 查看模版信息
	t, err := c.templateSvc.DetailTemplateByExternalTemplateId(ctx, evt.TemplateID)
	if err != nil {
		return fmt.Errorf("查看模版信息错误: %w", err)
	}

	// 查询用户信息
	wUser, err := c.userSvc.FindByWechatUser(ctx, evt.Applicant.UserID)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.logger.Error("创建工单，查询用户错误", elog.FieldErr(err))
		wUser.Username = evt.Applicant.UserID
	}

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// 如果用户不存在，使用申请人的UserID作为用户名
			c.logger.Info("未找到用户，使用申请人的UserID作为用户名",
				elog.String("user", evt.Applicant.UserID))
			wUser = user.User{
				Username: evt.Applicant.UserID,
			}
		} else {
			// 如果是其他错误，记录错误并返回
			c.logger.Error("创建工单，查询用户错误", elog.FieldErr(err))
			return err
		}
	}

	// 创建工单
	err = c.svc.CreateOrder(ctx, domain.Order{
		CreateBy:     wUser.Username,
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
