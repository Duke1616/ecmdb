package grpc

import (
	"context"

	orderv1 "github.com/Duke1616/ecmdb/api/proto/gen/order/v1"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/pkg/protox"

	"google.golang.org/grpc"
)

type WorkOrderServer struct {
	orderv1.UnimplementedWorkOrderServiceServer

	orderSvc service.Service
}

func NewWorkOrderServer(orderSvc service.Service) *WorkOrderServer {
	return &WorkOrderServer{orderSvc: orderSvc}
}

func (f *WorkOrderServer) Register(server grpc.ServiceRegistrar) {
	orderv1.RegisterWorkOrderServiceServer(server, f)
}

func (f *WorkOrderServer) CreateWorkOrder(ctx context.Context, request *orderv1.CreateOrderRequest) (
	*orderv1.Response, error) {
	// 解析工单数据
	data, err := protox.AnyMapToInterfaceMap(request.Order.Data)
	if err != nil {
		return nil, err
	}

	// 解析消息通知模版数据
	param, err := protox.AnyMapToInterfaceMap(request.NotificationConf.TemplateParams)
	if err != nil {
		return nil, err
	}

	// 创建工单
	err = f.orderSvc.CreateOrder(ctx, domain.Order{
		Provide:    domain.Provide(request.Order.Provider),
		TemplateId: request.Order.TemplateId,
		WorkflowId: request.Order.WorkflowId,
		Data:       data,
		CreateBy:   request.Order.CreateBy,
		NotificationConf: domain.NotificationConf{
			TemplateID:     request.NotificationConf.TemplateId,
			TemplateParams: param,
			Channel:        domain.Channel(request.NotificationConf.Channel.String()),
		},
	})

	return &orderv1.Response{}, err
}
