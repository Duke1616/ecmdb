package ioc

import (
	"fmt"

	tenantv1 "github.com/Duke1616/ecmdb/api/proto/gen/eiam/tenant/v1"
	pluginservice "github.com/Duke1616/ecmdb/internal/service/plugin"
	grpcpkg "github.com/Duke1616/etask/pkg/grpc"
	"github.com/Duke1616/etask/pkg/grpc/registry"
	"github.com/spf13/viper"
)

const eiamClientConfigKey = "grpc.client.eiam"

func ProvideBuiltinTenantConfig() (pluginservice.BuiltinTenantConfig, error) {
	return pluginservice.BuiltinTenantConfig{
		AccessKey: viper.GetString("plugin.builtin.tenant.access_key"),
		SecretKey: viper.GetString("plugin.builtin.tenant.secret_key"),
	}, nil
}

func NewBuiltinTenantProvider(cfg pluginservice.BuiltinTenantConfig, reg registry.Registry) (pluginservice.BuiltinTenantProvider, error) {
	client, err := newTenantServiceClient(reg)
	if err != nil {
		return nil, err
	}
	return pluginservice.NewBuiltinTenantProvider(cfg, client), nil
}

func newTenantServiceClient(reg registry.Registry) (tenantv1.TenantServiceClient, error) {
	clientCfg, err := loadEIAMClientConfig()
	if err != nil {
		return nil, err
	}

	conn, err := grpcpkg.NewClientConn(
		reg,
		grpcpkg.WithServiceName(clientCfg.Name),
		grpcpkg.WithClientJWTAuth(clientCfg.AuthToken),
	)
	if err != nil {
		return nil, fmt.Errorf("init eiam tenant grpc client: %w", err)
	}
	return tenantv1.NewTenantServiceClient(conn), nil
}

func loadEIAMClientConfig() (grpcpkg.ClientConfig, error) {
	var clientCfg grpcpkg.ClientConfig
	if err := viper.UnmarshalKey(eiamClientConfigKey, &clientCfg); err != nil {
		return grpcpkg.ClientConfig{}, fmt.Errorf("decode %s config: %w", eiamClientConfigKey, err)
	}
	return clientCfg, nil
}
