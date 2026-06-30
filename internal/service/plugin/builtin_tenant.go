package plugin

import (
	"context"
	"fmt"

	tenantv1 "github.com/Duke1616/ecmdb/api/proto/gen/eiam/tenant/v1"
	"github.com/Duke1616/eiam/pkg/ctxutil"
	"google.golang.org/grpc"
)

// BuiltinTenantConfig 描述内置插件导入时使用的租户配置。
type BuiltinTenantConfig struct {
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
}

// BuiltinTenantProvider 为内置插件导入命令提供带租户信息的 Context。
type BuiltinTenantProvider interface {
	// Context 返回已经注入租户 ID 的 Context。
	Context(ctx context.Context) (context.Context, error)
}

type tenantVerifier interface {
	// Verify 校验租户密钥并返回租户 ID。
	Verify(ctx context.Context, in *tenantv1.VerifyRequest, opts ...grpc.CallOption) (*tenantv1.VerifyResponse, error)
}

type builtinTenantProvider struct {
	cfg      BuiltinTenantConfig
	verifier tenantVerifier
}

func NewBuiltinTenantProvider(cfg BuiltinTenantConfig, verifier tenantVerifier) BuiltinTenantProvider {
	return &builtinTenantProvider{
		cfg:      cfg,
		verifier: verifier,
	}
}

func (p *builtinTenantProvider) Context(ctx context.Context) (context.Context, error) {
	tenantID, err := p.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	ctx = ctxutil.WithTenantID(ctx, tenantID)
	return ctxutil.WithOriginTenantID(ctx, tenantID), nil
}

func (p *builtinTenantProvider) tenantID(ctx context.Context) (int64, error) {
	if p.cfg.AccessKey == "" || p.cfg.SecretKey == "" {
		return 0, fmt.Errorf("内置插件租户配置不能为空 (请配置 --access-key 和 --secret-key 凭证以通过 EIAM 服务解析租户 ID)")
	}

	if p.verifier == nil {
		return 0, fmt.Errorf("内置插件租户校验客户端未初始化")
	}

	resp, err := p.verifier.Verify(ctx, &tenantv1.VerifyRequest{
		AccessKey: p.cfg.AccessKey,
		SecretKey: p.cfg.SecretKey,
	})
	if err != nil {
		return 0, fmt.Errorf("校验内置插件租户失败: %w", err)
	}
	if resp.GetTenantId() <= 0 {
		return 0, fmt.Errorf("校验内置插件租户失败: tenant_id 为空")
	}
	return resp.GetTenantId(), nil
}
