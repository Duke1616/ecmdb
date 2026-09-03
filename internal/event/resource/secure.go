package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Duke1616/ecmdb/internal/domain"
	resourceservice "github.com/Duke1616/ecmdb/internal/service/resource"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	mqx "github.com/Duke1616/ecmdb/pkg/mqx"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
	"github.com/samber/lo"
)

// secureTask 封装携带请求 Context 与租户信息的安全属性变更任务
type secureTask struct {
	ctx context.Context
	evt domain.FieldSecureAttrChange
}

// FieldSecureAttrChangeConsumer 负责在模型属性的安全状态（加密/解密）变更时，异步批量重加密资产历史数据
type FieldSecureAttrChangeConsumer struct {
	consumer     mq.Consumer
	svc          resourceservice.EncryptedSvc
	logger       *elog.Component
	workers      sync.Map
	idleDuration time.Duration
	limit        int64
	rootCtx      context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewFieldSecureAttrChangeConsumer 构造安全属性变更事件消费者
func NewFieldSecureAttrChangeConsumer(consumer mq.Consumer, svc resourceservice.EncryptedSvc, limit int64) *FieldSecureAttrChangeConsumer {
	return &FieldSecureAttrChangeConsumer{
		consumer:     consumer,
		svc:          svc,
		logger:       elog.DefaultLogger,
		workers:      sync.Map{},
		idleDuration: time.Minute * 10,
		limit:        limit,
	}
}

// Start 启动后台消费循环，支持响应上下文取消实现优雅停机
func (c *FieldSecureAttrChangeConsumer) Start(ctx context.Context) {
	c.rootCtx, c.cancel = context.WithCancel(ctx)
	for {
		select {
		case <-c.rootCtx.Done():
			c.logger.Info("模型安全属性字段变更消费者收到退出信号，停止消费")
			c.wg.Wait()
			return
		default:
			err := c.Consume(c.rootCtx)
			if err != nil {
				if c.rootCtx.Err() != nil {
					c.wg.Wait()
					return
				}
				c.logger.Error("模型安全属性字段变更，同步资产数据变更失败", elog.Any("错误信息", err))
				time.Sleep(time.Second)
			}
		}
	}
}

// Consume 提取单条消息并反序列化为事件
func (c *FieldSecureAttrChangeConsumer) Consume(ctx context.Context) error {
	// 使用 mqx.ConsumeMessage 恢复消息头（如 x-tenant-id）到 ctx
	ctxWithHeaders, cm, err := mqx.ConsumeMessage(ctx, c.consumer)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	var evt domain.FieldSecureAttrChange
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	return c.Process(ctxWithHeaders, evt)
}

// Process 调度针对指定属性的单协程任务流水线
func (c *FieldSecureAttrChangeConsumer) Process(ctx context.Context, evt domain.FieldSecureAttrChange) error {
	key := evt.ModelUid + ":" + evt.FieldUid

	chAny, loaded := c.workers.LoadOrStore(key, make(chan secureTask, 1))
	ch := chAny.(chan secureTask)

	if !loaded {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			c.runWorker(c.rootCtx, key, ch)
		}()
	}

	task := secureTask{
		ctx: ctx,
		evt: evt,
	}

	// 覆盖写入最新事件（保证 channel 中存的是最新任务，丢弃冗余重复的旧待办）
	select {
	case <-ch:
	default:
	}

	select {
	case ch <- task:
	case <-c.rootCtx.Done():
		return c.rootCtx.Err()
	}

	return nil
}

// runWorker 负责单一属性维度的串行异步批量更新，支持空闲回收与优雅退出
func (c *FieldSecureAttrChangeConsumer) runWorker(ctx context.Context, key string, ch chan secureTask) {
	idleTimer := time.NewTimer(c.idleDuration)
	defer idleTimer.Stop()

	for {
		select {
		case task, ok := <-ch:
			if !ok {
				return
			}

			// 组合上下文：继承任务专属上下文（包含链路追踪与租户 Headers），同时受组件全局 rootCtx 取消控制
			taskCtx, cancel := context.WithCancel(task.ctx)
			go func() {
				select {
				case <-ctx.Done():
					cancel()
				case <-taskCtx.Done():
				}
			}()

			if err := c.handleEvent(taskCtx, task.evt); err != nil {
				c.logger.Error("处理安全属性字段变更失败", elog.String("key", key), elog.Any("err", err))
			}
			cancel()

			if !idleTimer.Stop() {
				select {
				case <-idleTimer.C:
				default:
				}
			}
			idleTimer.Reset(c.idleDuration)

		case <-idleTimer.C:
			// 没有新事件，退出协程并从 workers 注册表中移除
			c.workers.Delete(key)
			return

		case <-ctx.Done():
			return
		}
	}
}

// handleEvent 执行具体的分页查询解密并批量重加密保存
func (c *FieldSecureAttrChangeConsumer) handleEvent(ctx context.Context, evt domain.FieldSecureAttrChange) error {
	var offset int64 = 0
	const maxLoops = 5000 // 看门狗防护，防止极端异常脏数据导致永久死循环

	for loop := 0; loop < maxLoops; loop++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		resources, err := c.svc.ListAndDecryptBeforeUtime(ctx, evt.TiggerTime, []string{evt.FieldUid},
			evt.ModelUid, offset, c.limit)
		if err != nil {
			return fmt.Errorf("field secure attr change list resources failed: %w", err)
		}
		if len(resources) == 0 {
			break
		}

		rs := lo.Map(resources, func(src domain.Resource, _ int) domain.Resource {
			return domain.Resource{
				ID:       src.ID,
				Name:     src.Name,
				ModelUID: src.ModelUID,
				Data: mongox.MapStr{
					evt.FieldUid: src.Data[evt.FieldUid],
				},
			}
		})

		modifiedCount, err := c.svc.BatchUpdateResources(ctx, rs)
		if err != nil {
			return fmt.Errorf("field secure attr change: batch update failed: %w", err)
		}

		// 防死循环保护：若当前批次未产生实际文档修改（例如空属性或 utime 未刷新），主动递增 offset 推进游标
		if modifiedCount == 0 {
			offset += int64(len(resources))
		}

		if int64(len(resources)) < c.limit {
			break
		}
	}
	return nil
}
