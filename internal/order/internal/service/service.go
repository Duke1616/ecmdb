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
	UpdateStatus(ctx context.Context, id int64, status uint8) error
	// RegisterProcessInstanceId 注册流程引擎ID
	RegisterProcessInstanceId(ctx context.Context, id int64, instanceId int) error
	// ListOrderByProcessEngineIds 获取代办流程
	ListOrderByProcessEngineIds(ctx context.Context, engineIds []int) (domain.Order, error)
}

type service struct {
	repo     repository.OrderRepository
	producer event.CreateProcessEventProducer
	l        *elog.Component
}

func (s *service) UpdateStatus(ctx context.Context, id int64, status uint8) error {
	//TODO implement me
	panic("implement me")
}

func (s *service) RegisterProcessInstanceId(ctx context.Context, id int64, instanceId int) error {
	return s.repo.RegisterProcessInstanceId(ctx, id, instanceId, domain.PROCESS.ToUint8())
}

func NewService(repo repository.OrderRepository, producer event.CreateProcessEventProducer) Service {
	return &service{
		repo:     repo,
		producer: producer,
		l:        elog.DefaultLogger,
	}
}

func (s *service) CreateOrder(ctx context.Context, req domain.Order) error {
	orderId, err := s.repo.CreateOrder(ctx, req)
	if err != nil {
		return err
	}

	go func() {
		err = s.sendGenerateFlowEvent(ctx, req, orderId)
	}()

	return err
}

func (s *service) ListOrderByProcessEngineIds(ctx context.Context, engineIds []int) (domain.Order, error) {
	return s.repo.ListOrderByProcessEngineIds(ctx, engineIds)
}

func (s *service) sendGenerateFlowEvent(ctx context.Context, req domain.Order, orderId int64) error {
	req.Data["starter"] = req.CreateBy
	evt := event.OrderEvent{
		Id:         orderId,
		WorkflowId: req.WorkflowId,
		Data:       req.Data,
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
