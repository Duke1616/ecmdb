package ioc

import (
	"fmt"

	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/spf13/viper"
)

// InitCrypto 初始化全局资产加解密与安全脱敏组件。
func InitCrypto() cryptox.Crypto {
	type Config struct {
		Version        string   `mapstructure:"version"`
		Key            string   `mapstructure:"key"`
		LegacyVersions []string `mapstructure:"legacy_versions"`
	}

	var cfg Config

	if err := viper.UnmarshalKey("encryption", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into structure: %v", err))
	}

	// 验证配置
	if cfg.Version == "" {
		panic(fmt.Errorf("encryption version is required"))
	}
	if cfg.Key == "" {
		panic(fmt.Errorf("encryption key is required"))
	}

	legacyVersions := cfg.LegacyVersions
	if len(legacyVersions) == 0 {
		legacyVersions = []string{cfg.Version}
	}

	// 构造并直接返回全局唯一加密与脱敏保护组件
	manager := cryptox.NewCryptoManager("V2").
		Register("V2", cryptox.MustNewAESCryptoV2(cfg.Key)).
		Register(cfg.Version, cryptox.MustNewAESCrypto(cfg.Key)).
		WithLegacyAlgos(legacyVersions...)

	return cryptox.NewValueProtector(manager)
}
