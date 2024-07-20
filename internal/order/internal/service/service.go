package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/event"
	"github.com/Duke1616/ecmdb/internal/order/internal/repository"
	"github.com/gotomicro/ego/core/elog"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	CreateOrder(ctx context.Context, req domain.Order) error
	DetailByProcessInstId(ctx context.Context, instanceId int) (domain.Order, error)
	UpdateStatusByInstanceId(ctx context.Context, instanceId int, status uint8) error
	// RegisterProcessInstanceId 注册流程引擎ID
	RegisterProcessInstanceId(ctx context.Context, id int64, instanceId int) error
	// ListOrderByProcessInstanceIds 获取代办流程
	ListOrderByProcessInstanceIds(ctx context.Context, instanceIds []int) ([]domain.Order, error)

	// ListHistoryOrder 获取历史order列表
	ListHistoryOrder(ctx context.Context, userId string, offset, limit int64) ([]domain.Order, int64, error)

	// ListOrdersByUser 查看自己提交的工单
	ListOrdersByUser(ctx context.Context, userId string, offset, limit int64) ([]domain.Order, int64, error)
}

type service struct {
	repo     repository.OrderRepository
	producer event.CreateProcessEventProducer
	l        *elog.Component
}

func (s *service) ListOrdersByUser(ctx context.Context, userId string, offset, limit int64) ([]domain.Order, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Order
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListOrder(ctx, userId, domain.PROCESS.ToUint8(), offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountOrder(ctx, userId, domain.PROCESS.ToUint8())
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) UpdateStatusByInstanceId(ctx context.Context, instanceId int, status uint8) error {
	return s.repo.UpdateStatusByInstanceId(ctx, instanceId, status)
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

func (s *service) ListOrderByProcessInstanceIds(ctx context.Context, instanceIds []int) ([]domain.Order, error) {
	return s.repo.ListOrderByProcessInstanceIds(ctx, instanceIds)
}

func (s *service) DetailByProcessInstId(ctx context.Context, instanceId int) (domain.Order, error) {
	return s.repo.DetailByProcessInstId(ctx, instanceId)
}

func (s *service) ListHistoryOrder(ctx context.Context, userId string, offset, limit int64) (
	[]domain.Order, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Order
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListOrder(ctx, userId, domain.END.ToUint8(), offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountOrder(ctx, userId, domain.END.ToUint8())
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) sendGenerateFlowEvent(ctx context.Context, req domain.Order, orderId int64) error {
	if req.Data == nil {
		req.Data = make(map[string]interface{})
	}

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
