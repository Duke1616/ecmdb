package repository

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type OrderRepository interface {
	CreateBizOrder(ctx context.Context, order domain.Order) (domain.Order, error)

	CreateOrder(ctx context.Context, req domain.Order) (int64, error)
	DetailByProcessInstId(ctx context.Context, instanceId int) (domain.Order, error)
	Detail(ctx context.Context, id int64) (domain.Order, error)
	RegisterProcessInstanceId(ctx context.Context, id int64, instanceId int, status uint8) error
	ListOrderByProcessInstanceIds(ctx context.Context, instanceIds []int) ([]domain.Order, error)
	UpdateStatusByInstanceId(ctx context.Context, instanceId int, status uint8) error
	ListOrder(ctx context.Context, userId string, status []int, offset, limit int64) ([]domain.Order, error)
	CountOrder(ctx context.Context, userId string, status []int) (int64, error)
	// FindByBizIdAndKey 根据 BizID 和 Key 查询工单
	FindByBizIdAndKey(ctx context.Context, bizId int64, key string, status []domain.Status) (domain.Order, error)
}

func NewOrderRepository(dao dao.OrderDAO) OrderRepository {
	return &orderRepository{
		dao: dao,
	}
}

type orderRepository struct {
	dao dao.OrderDAO
}

func (repo *orderRepository) CreateBizOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	resp, err := repo.dao.CreateBizOrder(ctx, repo.toEntity(order))
	return repo.toDomain(resp), err
}

func (repo *orderRepository) Detail(ctx context.Context, id int64) (domain.Order, error) {
	order, err := repo.dao.Detail(ctx, id)
	return repo.toDomain(order), err
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

func (repo *orderRepository) ListOrder(ctx context.Context, userId string, status []int, offset, limit int64) ([]domain.Order, error) {
	orders, err := repo.dao.ListOrder(ctx, userId, status, offset, limit)
	return slice.Map(orders, func(idx int, src dao.Order) domain.Order {
		return repo.toDomain(src)
	}), err
}

func (repo *orderRepository) CountOrder(ctx context.Context, userId string, status []int) (int64, error) {
	return repo.dao.CountOrder(ctx, userId, status)
}

func (repo *orderRepository) FindByBizIdAndKey(ctx context.Context, bizId int64, key string, status []domain.Status) (domain.Order, error) {
	statusUint8 := slice.Map(status, func(idx int, src domain.Status) uint8 {
		return src.ToUint8()
	})
	order, err := repo.dao.FindByBizIdAndKey(ctx, bizId, key, statusUint8)
	return repo.toDomain(order), err
}

func (repo *orderRepository) toEntity(req domain.Order) dao.Order {
	return dao.Order{
		BizID:      req.BizID,
		Key:        req.Key,
		TemplateId: req.TemplateId,
		Status:     req.Status.ToUint8(),
		Provide:    req.Provide.ToUint8(),
		WorkflowId: req.WorkflowId,
		CreateBy:   req.CreateBy,
		Data:       req.Data,
		NotificationConf: dao.NotificationConf{
			TemplateID:     req.NotificationConf.TemplateID,
			TemplateParams: req.NotificationConf.TemplateParams,
			Channel:        req.NotificationConf.Channel.String(),
		},
	}
}

func (repo *orderRepository) toDomain(req dao.Order) domain.Order {
	return domain.Order{
		Id:         req.Id,
		BizID:      req.BizID,
		Key:        req.Key,
		TemplateId: req.TemplateId,
		Status:     domain.Status(req.Status),
		Provide:    domain.Provide(req.Provide),
		WorkflowId: req.WorkflowId,
		Process:    domain.Process{InstanceId: req.ProcessInstanceId},
		CreateBy:   req.CreateBy,
		Data:       req.Data,
		Ctime:      req.Ctime,
		Wtime:      req.Wtime,
		NotificationConf: domain.NotificationConf{
			TemplateID:     req.NotificationConf.TemplateID,
			TemplateParams: req.NotificationConf.TemplateParams,
			Channel:        domain.Channel(req.NotificationConf.Channel),
		},
	}
}
