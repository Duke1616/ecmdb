//go:build wireinject

package ioc

import (
	"github.com/google/wire"
)

// InitApp 初始化完整应用
func InitApp() (*App, error) {
	wire.Build(
		WebSet,
		wire.Struct(new(App), "*"),
	)
	return nil, nil
}
