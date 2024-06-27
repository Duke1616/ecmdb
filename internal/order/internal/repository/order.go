package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/repository/dao"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, req domain.Order) (int64, error)
}

func NewOrderRepository(dao dao.OrderDAO) OrderRepository {
	return &orderRepository{
		dao: dao,
	}
}

type orderRepository struct {
	dao dao.OrderDAO
}

func (repo *orderRepository) CreateOrder(ctx context.Context, req domain.Order) (int64, error) {
	return repo.dao.CreateOrder(ctx, repo.toEntity(req))
}

func (repo *orderRepository) toEntity(req domain.Order) dao.Order {
	return dao.Order{
		Applicant: req.Applicant,
		Data:      req.Data,
	}
}
