package grpc

import (
	"context"
	"fmt"

	policyv1 "github.com/Duke1616/ecmdb/api/proto/gen/policy/v1"
	"github.com/Duke1616/ecmdb/internal/policy/internal/service"
	"google.golang.org/grpc"
)

type PolicyServer struct {
	policyv1.UnimplementedPolicyServiceServer

	policySvc service.Service
}

func NewPolicyServer(policySvc service.Service) *PolicyServer {
	return &PolicyServer{policySvc: policySvc}
}

func (f *PolicyServer) Register(server grpc.ServiceRegistrar) {
	policyv1.RegisterPolicyServiceServer(server, f)
}

func (f *PolicyServer) Authorize(ctx context.Context, request *policyv1.AuthorizeReq) (
	*policyv1.Response, error) {
	fmt.Println("Authorize", request.Resource)
	authorize, err := f.policySvc.Authorize(ctx, request.UserId, request.Path, request.Method, request.Resource)
	if err != nil {
		return &policyv1.Response{Allowed: false}, err
	}

	return &policyv1.Response{
		Allowed: authorize,
	}, err
}
