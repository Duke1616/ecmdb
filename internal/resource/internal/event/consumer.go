package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/service"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
)

const FieldSecureAttrChangeName = "field_secure_attr_change"

type FieldSecureAttrChangeConsumer struct {
	consumer mq.Consumer
	svc      service.EncryptedSvc
	logger   *elog.Component
	crypto   cryptox.Crypto[string]
	limit    int64
	offset   int64
}

func NewFieldSecureAttrChangeConsumer(q mq.MQ, svc service.EncryptedSvc, limit int64,
	crypto cryptox.Crypto[string]) (*FieldSecureAttrChangeConsumer, error) {
	consumer, err := q.Consumer(FieldSecureAttrChangeName, "field_secure_change")
	if err != nil {
		return nil, fmt.Errorf("获取消息失败: %w", err)
	}
	return &FieldSecureAttrChangeConsumer{
		consumer: consumer,
		svc:      svc,
		logger:   elog.DefaultLogger,
		limit:    limit,
		crypto:   crypto,
		offset:   0,
	}, nil
}

func (c *FieldSecureAttrChangeConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				c.logger.Error("模型安全属性字段变更，同步资产数据变更失败", elog.Any("错误信息", err))
			}
		}
	}()
}

func (c *FieldSecureAttrChangeConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	var evt attribute.Event
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	return c.Process(ctx, evt)
}

func (c *FieldSecureAttrChangeConsumer) Process(ctx context.Context, evt attribute.Event) error {
	for {
		resources, err := c.svc.ListBeforeUtime(ctx, evt.TiggerTime, []string{evt.FieldUid},
			evt.ModelUid, c.offset, c.limit)
		if err != nil {
			return fmt.Errorf("field secure attr change list resources failed: %w", err)
		}
		if len(resources) == 0 {
			break
		}

		rs := slice.Map(resources, func(idx int, src domain.Resource) domain.Resource {
			return domain.Resource{
				ID:       src.ID,
				Name:     src.Name,
				ModelUID: src.ModelUID,
				Data: mongox.MapStr{
					evt.FieldUid: src.Data[evt.FieldUid],
				},
			}
		})

		if _, err = c.svc.BatchUpdateResources(ctx, rs); err != nil {
			return fmt.Errorf("field secure attr change: batch update failed: %w", err)
		}

		// 如果不足一页，说明到末尾了
		if int64(len(resources)) < c.limit {
			break
		}
	}

	return nil
}
