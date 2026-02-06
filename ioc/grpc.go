package ioc

import (
	"github.com/Duke1616/ecmdb/internal/endpoint"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/rota"
	"github.com/Duke1616/ecmdb/internal/user"

	grpcpkg "github.com/Duke1616/ework-runner/pkg/grpc"
	registrysdk "github.com/Duke1616/ework-runner/pkg/grpc/registry"

	"github.com/spf13/viper"
)

func InitGrpcServer(registry registrysdk.Registry, orderRpc *order.RpcServer, policyRpc *policy.RpcServer,
	endpointRpc *endpoint.RpcServer, userRpc *user.RpcServer, rotaRpc *rota.RpcServer) *grpcpkg.Server {
	var cfg grpcpkg.ServerConfig
	if err := viper.UnmarshalKey("grpc.server.ecmdb", &cfg); err != nil {
		panic(err)
	}

	server := grpcpkg.NewServer(cfg, registry, grpcpkg.WithJWTAuth(cfg.AuthToken))

	orderRpc.Register(server)
	policyRpc.Register(server)
	endpointRpc.Register(server)
	userRpc.Register(server)
	rotaRpc.Register(server)

	return server
}
