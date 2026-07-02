package ioc

import (
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/Duke1616/eiam/pkg/web/sdk"
)

// InitPolicySDK 初始化 EIAM 鉴权 SDK
func InitPolicySDK() *sdk.SDK {
	return sdk.NewSDK()
}

// InitPermSyncer 初始化 EIAM 资产上报同步器
func InitPermSyncer() capability.Syncer {
	return capability.NewSyncer(capability.NewHttpReporter())
}

// InitProviders 提供逻辑权限资源列表（此处默认为 nil，依赖自动发现）
func InitProviders() []capability.PermissionProvider {
	return nil
}
