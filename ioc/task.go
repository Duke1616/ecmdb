package ioc

import (
	"time"

	executorv1 "github.com/Duke1616/ecmdb/api/proto/gen/etask/executor/v1"
	taskv1 "github.com/Duke1616/ecmdb/api/proto/gen/etask/task/v1"
	grpcpkg "github.com/Duke1616/ework-runner/pkg/grpc"
	"github.com/Duke1616/ework-runner/pkg/grpc/registry"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// TaskClientConn 用于区分 TASK gRPC 客户端连接
type TaskClientConn struct {
	grpc.ClientConnInterface
}

// InitTASKGrpcClient 初始化 ETASK gRPC 客户端
func InitTASKGrpcClient(reg registry.Registry) TaskClientConn {
	var cfg grpcpkg.ClientConfig
	// 这里之前写错了 ealert，修复为 etask
	if err := viper.UnmarshalKey("grpc.client.etask", &cfg); err != nil {
		panic(err)
	}

	// 通过 WaitForReady 控制，如果地址不通直接返回错误
	cc, err := grpcpkg.NewClientConn(
		reg,
		grpcpkg.WithServiceName(cfg.Name),
		grpcpkg.WithClientJWTAuth(cfg.AuthToken),
		grpcpkg.WithDialOption(grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 3 * time.Second,
		}),
			grpc.WithDefaultCallOptions(
				grpc.WaitForReady(false),
			)),
	)
	if err != nil {
		panic(err)
	}

	return TaskClientConn{cc}
}

func InitTaskServiceClient(cc TaskClientConn) taskv1.TaskServiceClient {
	return taskv1.NewTaskServiceClient(cc)
}

func InitTaskExecutionServiceClient(cc TaskClientConn) executorv1.TaskExecutionServiceClient {
	return executorv1.NewTaskExecutionServiceClient(cc)
}
