package ioc

import (
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/pkg/grpcx"
	"github.com/Duke1616/ecmdb/pkg/grpcx/interceptors/jwt"
	"github.com/gotomicro/ego/core/elog"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func InitGrpcServer(orderRpc *order.RpcServer, policyRpc *policy.RpcServer, etcdClient *clientv3.Client) *grpcx.Server {
	type Config struct {
		Name    string `mapstructure:"name"`
		Port    int    `mapstructure:"port"`
		EtcdTTL int64  `mapstructure:"etcd_ttl"`
		Key     string `mapstructure:"key"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}

	if cfg.Name == "" {
		panic("server name cannot be empty")
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		panic("port must be between 1 and 65535, got: %d")
	}

	if cfg.EtcdTTL < 1 {
		panic("etcd TTL must be greater than 0")
	}

	if cfg.Key == "" {
		panic("JWT key cannot be empty")
	}

	jwtInterceptor := jwt.NewJwtAuth(cfg.Key)
	server := grpc.NewServer(grpc.UnaryInterceptor(
		jwtInterceptor.JwtAuthInterceptor(),
	))
	orderRpc.Register(server)
	policyRpc.Register(server)

	return &grpcx.Server{
		Server:     server,
		Port:       cfg.Port,
		Name:       cfg.Name,
		L:          elog.DefaultLogger,
		EtcdClient: etcdClient,
		EtcdTTL:    cfg.EtcdTTL,
	}
}
