package startup

import (
	"fmt"

	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/spf13/viper"
)

func InitCryptoRegistry() *cryptox.CryptoRegistry {
	type EncryptionConfig struct {
		Version string `mapstructure:"version"`
		Key     string `mapstructure:"key"`
	}

	type Config struct {
		User     EncryptionConfig `mapstructure:"user"`
		Resource EncryptionConfig `mapstructure:"resource"`
		Runner   EncryptionConfig `mapstructure:"runner"`
	}

	var cfg Config

	if err := viper.UnmarshalKey("encryption", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into structure: %v", err))
	}

	// 验证配置
	validateConfig := func(ec EncryptionConfig, name string) {
		if ec.Version == "" {
			panic(fmt.Errorf("%s encryption version is required", name))
		}
		if ec.Key == "" {
			panic(fmt.Errorf("%s encryption key is required", name))
		}
	}

	validateConfig(cfg.User, "user")
	validateConfig(cfg.Resource, "resource")
	validateConfig(cfg.Runner, "runner")

	// 进行注册
	reg := &cryptox.CryptoRegistry{
		User: cryptox.NewCryptoManager[string](cfg.User.Version).
			RegisterAesAlgorithm(cfg.User.Version, cfg.User.Key).
			WithLegacyAlgo(cfg.User.Version),
		Resource: cryptox.NewCryptoManager[string](cfg.Resource.Version).
			RegisterAesAlgorithm(cfg.Resource.Version, cfg.Resource.Key).
			WithLegacyAlgo(cfg.User.Version),
		Runner: cryptox.NewCryptoManager[string](cfg.Runner.Version).
			RegisterAesAlgorithm(cfg.Runner.Version, cfg.Runner.Key).
			WithLegacyAlgo(cfg.User.Version),
	}

	return reg
}
