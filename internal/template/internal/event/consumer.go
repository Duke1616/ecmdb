// Copyright 2023 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package event

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
	"github.com/Duke1616/ecmdb/internal/template/internal/service"
	"log/slog"

	"github.com/ecodeclub/mq-api"
	"github.com/xen0n/go-workwx"
)

type WechatApprovalCallbackConsumer struct {
	svc      service.Service
	consumer mq.Consumer
}

func NewWechatApprovalCallbackConsumer(svc service.Service, q mq.MQ) (*WechatApprovalCallbackConsumer, error) {
	groupID := "callback"
	consumer, err := q.Consumer(CallbackEventName, groupID)
	if err != nil {
		return nil, err
	}
	return &WechatApprovalCallbackConsumer{
		svc:      svc,
		consumer: consumer,
	}, nil
}

func (c *WechatApprovalCallbackConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				slog.Error("同步事件失败", err)
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
		slog.Error("模版已经存在或新增模版失败", err)
	}

	return err
}

func (c *WechatApprovalCallbackConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}
