package grpc

import (
	"context"

	endpointv1 "github.com/Duke1616/ecmdb/api/proto/gen/endpoint/v1"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/domain"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/service"
	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type EndpointServer struct {
	endpointv1.UnimplementedEndpointServiceServer

	endpointSvc service.Service
}

func NewEndpointServer(endpointSvc service.Service) *EndpointServer {
	return &EndpointServer{endpointSvc: endpointSvc}
}

func (f *EndpointServer) Register(server grpc.ServiceRegistrar) {
	endpointv1.RegisterEndpointServiceServer(server, f)
}

func (f *EndpointServer) BatchRegister(ctx context.Context, request *endpointv1.BatchRegisterEndpointsReq) (
	*emptypb.Empty, error) {
	_, err := f.endpointSvc.BatchRegisterByResource(ctx, request.Resource, slice.Map(request.Endpoints,
		func(idx int, src *endpointv1.Endpoint) domain.Endpoint {
			return domain.Endpoint{
				Path:         src.Path,
				Method:       src.Method,
				Resource:     src.Resource,
				Desc:         src.Desc,
				IsAuth:       src.IsAuth,
				IsAudit:      src.IsAudit,
				IsPermission: src.IsPermission,
			}
		}))
	if err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, err
}
