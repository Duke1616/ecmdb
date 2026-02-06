package grpc

import (
	"context"

	orderv1 "github.com/Duke1616/ecmdb/api/proto/gen/order/v1"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/pkg/protox"

	"google.golang.org/grpc"
)

type WorkOrderServer struct {
	orderv1.UnimplementedWorkOrderServiceServer

	templateSvc template.Service
	orderSvc    service.Service
}

func NewWorkOrderServer(orderSvc service.Service, templateSvc template.Service) *WorkOrderServer {
	return &WorkOrderServer{
		orderSvc:    orderSvc,
		templateSvc: templateSvc,
	}
}

func (f *WorkOrderServer) Register(server grpc.ServiceRegistrar) {
	orderv1.RegisterWorkOrderServiceServer(server, f)
}

func (f *WorkOrderServer) CreateWorkOrder(ctx context.Context, request *orderv1.CreateOrderRequest) (
	*orderv1.Response, error) {
	// 解析工单数据
	data, err := protox.AnyMapToInterfaceMap(request.Data)
	if err != nil {
		return nil, err
	}

	// 获取模版信息
	tmInfo, err := f.templateSvc.DetailTemplate(ctx, request.TemplateId)
	if err != nil {
		return nil, err
	}

	// 构建结构体
	orderReq := domain.Order{
		Provide:    domain.Provide(request.Provider),
		TemplateId: tmInfo.Id,
		WorkflowId: tmInfo.WorkflowId,
		Data:       data,
		CreateBy:   request.CreateBy,
	}

	// TODO 如果外部传递为空，应该使用模版的 CreateBy 用户
	if request.CreateBy == "" {

	}

	// 创建工单
	// 是否根据考虑在这个地方直接消息通知？？？ 因为工单有创建失败的可能
	_, err = f.orderSvc.CreateBizOrder(ctx, orderReq)
	return &orderv1.Response{}, err
}
