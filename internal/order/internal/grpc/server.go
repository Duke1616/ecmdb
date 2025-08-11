package grpc

import (
	"context"
	orderv1 "github.com/Duke1616/ecmdb/api/proto/gen/order/v1"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"

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
	err := f.orderSvc.CreateOrder(ctx, domain.Order{
		Provide:    domain.Provide(request.Order.Provider),
		TemplateId: request.Order.TemplateId,
		WorkflowId: request.Order.WorkflowId,
	})

	return &orderv1.Response{}, err
}
