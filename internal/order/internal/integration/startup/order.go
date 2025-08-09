package startup

import (
	"context"
	jwtpkg "github.com/Duke1616/ecmdb/pkg/grpcx/interceptors/jwt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func InitEcmdbClient(etcdClient *etcdv3.Client) *grpc.ClientConn {
	type Config struct {
		Target string `json:"target"`
		Secure bool   `json:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.ecmdb", &cfg)
	if err != nil {
		panic(err)
	}
	rs, err := resolver.NewBuilder(etcdClient)
	if err != nil {
		panic(err)
	}
	opts := []grpc.DialOption{grpc.WithResolvers(rs)}
	if !cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.NewClient(cfg.Target, opts...)
	if err != nil {
		panic(err)
	}

	return cc
}

func contextWithJWT(ctx context.Context) context.Context {
	// 使用项目已有的JWT包创建令牌
	jwtAuth := jwtpkg.NewJwtAuth("test_key")

	// 创建包含业务ID的声明
	claims := jwt.MapClaims{
		"biz_id": float64(1),
	}

	// 使用JWT认证包的Encode方法生成令牌
	tokenString, _ := jwtAuth.Encode(claims)

	// 创建带有授权信息的元数据
	md := metadata.New(map[string]string{
		"Authorization": "Bearer " + tokenString,
	})
	return metadata.NewOutgoingContext(ctx, md)
}
