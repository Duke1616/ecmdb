package event

import (
	"context"

	"github.com/Duke1616/ecmdb/pkg/mqx"
	"github.com/ecodeclub/mq-api"
)

// TaskWorkerEventProducer 定义了工作节点事件生产者的接口
// 支持向多个 Topic 发送事件，并动态管理这些 Topic 的生产者资源
type TaskWorkerEventProducer interface {
	// Produce 将 Agent 执行事件发送到指定的 Topic
	Produce(ctx context.Context, topic string, evt AgentExecuteEvent) error

	// AddProducer 动态创建一个新的 Topic 生产者并缓存
	AddProducer(topic string) error

	// DelProducer 移除并关闭指定 Topic 的生产者资源
	DelProducer(topic string) error
}

func NewTaskRunnerEventProducer(q mq.MQ) (TaskWorkerEventProducer, error) {
	return mqx.NewMultipleProducer[AgentExecuteEvent](q)
}
