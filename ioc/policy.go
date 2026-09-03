package ioc

import (
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/Duke1616/eiam/pkg/web/capability/syncer"
	httpreporter "github.com/Duke1616/eiam/pkg/web/capability/syncer/http"
	"github.com/Duke1616/eiam/pkg/web/sdk"
)

// InitPolicySDK 初始化 EIAM 鉴权 SDK
func InitPolicySDK() *sdk.SDK {
	return sdk.NewSDK()
}

// InitPermSyncer 初始化资产同步器。
// httpreporter.New() 自动从 viper 读取 policy.discovery_url 与 policy.discovery_token，
// 若配置了 token 则自动注入 Authorization: Bearer 头，无需额外手动配置。
func InitPermSyncer() syncer.Syncer {
	return syncer.New(httpreporter.New())
}

// InitProviders 提供逻辑权限资源列表（此处默认为 nil，依赖自动发现）
func InitProviders() []capability.PermissionProvider {
	return nil
}

