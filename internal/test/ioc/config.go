package ioc

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func init() {
	// 自动寻径加载配置文件
	// 从当前测试执行目录向上查找，直到找到 internal/test/config.yaml
	curr, err := os.Getwd()
	if err != nil {
		return
	}

	for {
		configPath := filepath.Join(curr, "internal/test/config.yaml")
		if _, err = os.Stat(configPath); err == nil {
			viper.SetConfigFile(configPath)
			_ = viper.ReadInConfig()
			return
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			// 已到达根目录
			break
		}
		curr = parent
	}
}
