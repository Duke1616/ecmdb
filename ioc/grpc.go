package ioc

import (
	pluginv1 "github.com/Duke1616/ecmdb/api/proto/gen/ecmdb/plugin/v1"
	pluginserver "github.com/Duke1616/ecmdb/internal/grpc/plugin"
	grpcpkg "github.com/Duke1616/etask/pkg/grpc"
	registrysdk "github.com/Duke1616/etask/pkg/grpc/registry"
	"github.com/spf13/viper"
)

// InitGrpcServer 初始化 ecmdb 内部的 gRPC 服务端，并注册 pluginv1 服务
func InitGrpcServer(
	registry registrysdk.Registry,
	pluginServer *pluginserver.Server,
) *grpcpkg.Server {
	var cfg grpcpkg.ServerConfig
	if err := viper.UnmarshalKey("grpc.server.ecmdb", &cfg); err != nil {
		panic(err)
	}

	server := grpcpkg.NewServer(cfg, registry)

	pluginv1.RegisterPluginRuntimeServiceServer(server, pluginServer)
	return server
}
