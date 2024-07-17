package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, req domain.Order) (int64, error)
	DetailByProcessInstId(ctx context.Context, instanceId int) (domain.Order, error)
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

func (repo *orderRepository) DetailByProcessInstId(ctx context.Context, instanceId int) (domain.Order, error) {
	order, err := repo.dao.DetailByProcessInstId(ctx, instanceId)
	return repo.toDomain(order), err
}

func (repo *orderRepository) ListOrderByProcessInstanceIds(ctx context.Context, instanceIds []int) ([]domain.Order, error) {
	orders, err := repo.dao.ListOrderByProcessInstanceIds(ctx, instanceIds)
	return slice.Map(orders, func(idx int, src dao.Order) domain.Order {
		return repo.toDomain(src)
	}), err
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
		Process:    domain.Process{InstanceId: req.ProcessInstanceId},
		CreateBy:   req.CreateBy,
		Data:       req.Data,
		Ctime:      req.Ctime,
	}
}
