package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/runner/domain"
	"github.com/Duke1616/ecmdb/internal/runner/event"
	"log/slog"
)

type Service interface {
	CreateProducer(topic string) error
	DeleteProducer(topic string) error
	Publish(ctx context.Context, req domain.Runner)
}

type service struct {
	producer event.TaskRunnerEventProducer
}

func NewService(producer event.TaskRunnerEventProducer) Service {
	return &service{
		producer: producer,
	}
}
func (s *service) CreateProducer(topic string) error {
	return s.producer.AddProducer(topic)
}

func (s *service) DeleteProducer(topic string) error {
	return s.producer.DelProducer(topic)
}

func (s *service) Publish(ctx context.Context, req domain.Runner) {
	evt := event.RunnerEvent{
		Language: req.Language,
		Code:     req.Code,
		Name:     req.Name,
		UUID:     req.UUID,
	}

	if err := s.producer.Produce(ctx, req.Topic, evt); err != nil {
		slog.Error("工作节点发送指令失败",
			slog.Any("error", err),
			slog.Any("event", evt),
		)
	}
}
