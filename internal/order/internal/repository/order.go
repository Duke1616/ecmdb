package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/repository/dao"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, req domain.Order) (int64, error)
	RegisterProcessInstanceId(ctx context.Context, id int64, instanceId int, status uint8) error
	ListOrderByProcessInstanceIds(ctx context.Context, instanceIds []int) ([]domain.Order, error)
	UpdateStatusByInstanceId(ctx context.Context, instanceId int, status uint8) error
}

func NewOrderRepository(dao dao.OrderDAO) OrderRepository {
	return &orderRepository{
		dao: dao,
	}
}

type orderRepository struct {
	dao dao.OrderDAO
}

func (repo *orderRepository) UpdateStatusByInstanceId(ctx context.Context, instanceId int, status uint8) error {
	return repo.dao.UpdateStatusByInstanceId(ctx, instanceId, status)
}

func (repo *orderRepository) CreateOrder(ctx context.Context, req domain.Order) (int64, error) {
	return repo.dao.CreateOrder(ctx, repo.toEntity(req))
}

func (repo *orderRepository) RegisterProcessInstanceId(ctx context.Context, id int64, instanceId int, status uint8) error {
	return repo.dao.RegisterProcessInstanceId(ctx, id, instanceId, status)
}

func (repo *orderRepository) ListOrderByProcessInstanceIds(ctx context.Context, instanceIds []int) ([]domain.Order, error) {
	return nil, nil
}

func (repo *orderRepository) toEntity(req domain.Order) dao.Order {
	return dao.Order{
		TemplateId: req.TemplateId,
		Status:     req.Status.ToUint8(),
		WorkflowId: req.WorkflowId,
		CreateBy:   req.CreateBy,
		Data:       req.Data,
	}
}

func (repo *orderRepository) toDomain(req dao.Order) domain.Order {
	return domain.Order{
		TemplateId: req.TemplateId,
		Status:     domain.Status(req.Status),
		WorkflowId: req.WorkflowId,
		CreateBy:   req.CreateBy,
		Data:       req.Data,
	}
}
