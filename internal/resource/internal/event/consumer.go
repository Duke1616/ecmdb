package event

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

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
	consumer     mq.Consumer
	svc          service.EncryptedSvc
	logger       *elog.Component
	crypto       cryptox.Crypto[string]
	workers      sync.Map
	idleDuration time.Duration
	limit        int64
	offset       int64
}

func NewFieldSecureAttrChangeConsumer(q mq.MQ, svc service.EncryptedSvc, limit int64,
	crypto cryptox.Crypto[string]) (*FieldSecureAttrChangeConsumer, error) {
	consumer, err := q.Consumer(FieldSecureAttrChangeName, "field_secure_change")
	if err != nil {
		return nil, fmt.Errorf("获取消息失败: %w", err)
	}
	return &FieldSecureAttrChangeConsumer{
		consumer:     consumer,
		svc:          svc,
		logger:       elog.DefaultLogger,
		workers:      sync.Map{},
		idleDuration: time.Minute * 10,
		limit:        limit,
		crypto:       crypto,
		offset:       0,
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
	key := evt.ModelUid + ":" + evt.FieldUid

	chAny, loaded := c.workers.LoadOrStore(key, make(chan attribute.Event, 1))
	ch := chAny.(chan attribute.Event)

	if !loaded {
		go c.runWorker(ctx, key, ch)
	}

	// 覆盖写入最新事件（如果 channel 已经有值，先清掉）
	select {
	case <-ch:
	default:
	}
	ch <- evt

	return nil
}

func (c *FieldSecureAttrChangeConsumer) runWorker(ctx context.Context, key string, ch chan attribute.Event) {
	idleTimer := time.NewTimer(c.idleDuration)
	defer idleTimer.Stop()

	for {
		select {
		case evt := <-ch:
			if err := c.handleEvent(ctx, evt); err != nil {
				c.logger.Error("处理安全属性字段变更失败", elog.String("key", key), elog.Any("err", err))
			}
			if !idleTimer.Stop() {
				<-idleTimer.C
			}
			idleTimer.Reset(c.idleDuration)

		case <-idleTimer.C:
			// 没有新事件，退出并清理
			c.workers.Delete(key)
			close(ch)
			return
		}
	}
}

func (c *FieldSecureAttrChangeConsumer) handleEvent(ctx context.Context, evt attribute.Event) error {
	for {
		// 无论修改状态是加密或解密，都进行一次解密处理，BatchUpdate会在进行进行加密解密的处理
		resources, err := c.svc.ListAndDecryptBeforeUtime(ctx, evt.TiggerTime, []string{evt.FieldUid},
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

		if int64(len(resources)) < c.limit {
			break
		}
	}
	return nil
}
