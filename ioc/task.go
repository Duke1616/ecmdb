package ioc

import (
	taskv1 "github.com/Duke1616/ecmdb/api/proto/gen/task/v1"
	grpcpkg "github.com/Duke1616/ework-runner/pkg/grpc"
	"github.com/Duke1616/ework-runner/pkg/grpc/registry"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// InitETASKGrpcClient 初始化 ETASK gRPC 客户端
func InitETASKGrpcClient(reg registry.Registry) grpc.ClientConnInterface {
	var cfg grpcpkg.ClientConfig
	if err := viper.UnmarshalKey("grpc.client.ecmdb", &cfg); err != nil {
		panic(err)
	}

	cc, err := grpcpkg.NewClientConn(
		reg,
		grpcpkg.WithServiceName(cfg.Name),
		grpcpkg.WithClientJWTAuth(cfg.AuthToken),
	)
	if err != nil {
		panic(err)
	}

	return cc
}

// InitTaskServiceClient 初始化 Policy 服务客户端
func InitTaskServiceClient(cc grpc.ClientConnInterface) taskv1.TaskServiceClient {
	return taskv1.NewTaskServiceClient(cc)
}
