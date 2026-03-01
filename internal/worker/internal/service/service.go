package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/worker/internal/domain"
	"github.com/Duke1616/ecmdb/internal/worker/internal/event"
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
	kafkago "github.com/segmentio/kafka-go"
)

type Service interface {
	// EnsureInfrastructures 确保工作节点所需的基础设施就绪
	// 包括创建 Kafka Topic 以及初始化对应的消息生产者
	EnsureInfrastructures(ctx context.Context, topic string) error

	// Execute 将执行指令发送至指定的 Kafka Topic
	Execute(ctx context.Context, req domain.Execute) error

	// Release 释放指定 Topic 占用的基础设施资源
	// 当该 Topic 下不再有活跃的工作节点时调用，用于回收生产者资源
	Release(ctx context.Context, topic string) error
}

type service struct {
	logger   *elog.Component
	producer event.TaskWorkerEventProducer
	mq       mq.MQ
}

func (s *service) Execute(ctx context.Context, req domain.Execute) error {
	evt := event.AgentExecuteEvent{
		Language:  req.Language,
		Topic:     req.Topic,
		Handler:   req.Handler,
		Code:      req.Code,
		TaskId:    req.TaskId,
		Args:      req.Args,
		Variables: req.Variables,
	}

	err := s.producer.Produce(ctx, req.Topic, evt)
	if err != nil {
		s.logger.Debug("工作节点发送指令失败",
			elog.FieldErr(err),
			elog.Any("event", evt),
		)
	}

	return err
}

func NewService(mq mq.MQ, producer event.TaskWorkerEventProducer) Service {
	return &service{
		mq:       mq,
		logger:   elog.DefaultLogger,
		producer: producer,
	}
}

func (s *service) EnsureInfrastructures(ctx context.Context, topic string) error {
	// 新增 Topic，如果 Topic 已存在，mq.CreateTopic 可能会返回错误，这里选择判断如果是已存在则继续
	if err := s.mq.CreateTopic(ctx, topic, 1); err != nil {
		var val kafkago.Error
		if !errors.As(err, &val) || !errors.Is(val, kafkago.TopicAlreadyExists) {
			return fmt.Errorf("创建Topic失败: %x", err)
		}
	}

	// 新增 producer 监听
	if err := s.producer.AddProducer(topic); err != nil {
		return fmt.Errorf("推送 Producer 初始化失败: %x", err)
	}

	return nil
}

func (s *service) Release(ctx context.Context, topic string) error {
	return s.producer.DelProducer(topic)
}
