//go:build wireinject

package ioc

import (
	"github.com/google/wire"
)

// InitApp 初始化插件导入命令应用。
func InitApp() (*App, error) {
	wire.Build(
		PluginCommandSet,
		wire.Struct(new(App), "*"),
	)
	return nil, nil
}
