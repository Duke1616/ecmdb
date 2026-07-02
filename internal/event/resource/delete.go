package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/domain"
	resourceservice "github.com/Duke1616/ecmdb/internal/service/resource"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
)

// FieldDeleteConsumer 字段删除事件消费者
type FieldDeleteConsumer struct {
	consumer mq.Consumer
	svc      resourceservice.Service
	logger   *elog.Component
}

// NewFieldDeleteConsumer 构造字段删除消费者
func NewFieldDeleteConsumer(consumer mq.Consumer, svc resourceservice.Service) *FieldDeleteConsumer {
	return &FieldDeleteConsumer{
		consumer: consumer,
		svc:      svc,
		logger:   elog.DefaultLogger,
	}
}

// Start 启动后台消费协程
func (c *FieldDeleteConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				c.logger.Error("字段删除级联清理，同步资产数据变更失败", elog.Any("错误信息", err))
				time.Sleep(time.Second)
			}
		}
	}()
}

// Consume 提取事件消息并反序列化
func (c *FieldDeleteConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	var evt domain.FieldDelete
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	return c.Process(ctx, evt)
}

// Process 处理级联删除事件逻辑
func (c *FieldDeleteConsumer) Process(ctx context.Context, evt domain.FieldDelete) error {
	c.logger.Info("开始执行 Cascade Cleaner: 级联抹除模型平铺废弃属性",
		elog.String("model_uid", evt.ModelUid),
		elog.String("field_uid", evt.FieldUid))

	// 调用 service 层的 UnsetCustomField 抹除该 ModelUid 下所有实例的对应平铺字段
	modifiedCount, err := c.svc.UnsetCustomField(ctx, evt.ModelUid, evt.FieldUid)
	if err != nil {
		return fmt.Errorf("Cascade Cleaner 执行失败: %w", err)
	}

	c.logger.Info("Cascade Cleaner 级联抹除废弃属性执行成功",
		elog.String("model_uid", evt.ModelUid),
		elog.String("field_uid", evt.FieldUid),
		elog.Int64("modified_count", modifiedCount))

	return nil
}
