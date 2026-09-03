package mqx

import (
	"context"
	"time"
)

const handlerRetryDelay = time.Second

// ConsumeLoop 运行事件消费者循环。
// ResilientConsumer 负责 MQ 连接恢复；这里仅负责业务处理失败后的限速重试，
// 并保证服务停止时不会继续等待或再次调用业务处理函数。
func ConsumeLoop(ctx context.Context, consume func(context.Context) error, onError func(error)) {
	for {
		if ctx.Err() != nil {
			return
		}
		if err := consume(ctx); err == nil {
			continue
		} else {
			if ctx.Err() != nil {
				return
			}
			onError(err)
		}
		if !waitContext(ctx, handlerRetryDelay) {
			return
		}
	}
}

func waitContext(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
