package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/order"
)

type Service interface {
	// CreateTask 创建自动化任务
	CreateTask(ctx context.Context)
}

type service struct {
	orderSvc order.Service
}

func NewService(orderSvc order.Service) Service {
	return &service{
		orderSvc: orderSvc,
	}
}

func (s *service) CreateTask(ctx context.Context) {
	//TODO implement me
	panic("implement me")
}
