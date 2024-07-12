package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/event"
	"github.com/Duke1616/ecmdb/internal/order/internal/repository"
	"github.com/gotomicro/ego/core/elog"
)

type Service interface {
	CreateOrder(ctx context.Context, req domain.Order) error
}

type service struct {
	repo     repository.OrderRepository
	producer event.CreateFlowEventProducer
	l        *elog.Component
}

func NewService(repo repository.OrderRepository, producer event.CreateFlowEventProducer) Service {
	return &service{
		repo:     repo,
		producer: producer,
		l:        elog.DefaultLogger,
	}
}

func (s *service) CreateOrder(ctx context.Context, req domain.Order) error {
	_, err := s.repo.CreateOrder(ctx, req)
	if err != nil {
		return err
	}

	return s.sendGenerateFlowEvent(ctx, req)
}

func (s *service) sendGenerateFlowEvent(ctx context.Context, req domain.Order) error {
	evt := event.OrderEvent{
		FlowId: req.FlowId,
		Data:   req.Data,
	}
	err := s.producer.Produce(ctx, evt)

	if err != nil {
		// 要做好监控和告警
		s.l.Error("发送创建流程事件失败",
			elog.FieldErr(err),
			elog.Any("evt", evt))
	}

	return nil
}
