package service

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/strategy/internal/domain"
)

type Service interface {
	CreateOrder(ctx context.Context, req domain.Strategy) (int64, error)
}

type service struct {
}

func NewService() Service {
	return &service{}
}

func (s *service) CreateOrder(ctx context.Context, req domain.Strategy) (int64, error) {
	// TODO 匹配相应的规则, 去创建工单

	// TODO 去除重复
	panic("implement me")
}
