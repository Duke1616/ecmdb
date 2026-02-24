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
		User: cryptox.NewCryptoManager("V2").
			Register("V2", cryptox.MustNewAESCryptoV2(cfg.User.Key)).
			Register(cfg.User.Version, cryptox.MustNewAESCrypto(cfg.User.Key)).
			WithLegacyAlgo(cfg.User.Version),
		Resource: cryptox.NewCryptoManager("V2").
			Register("V2", cryptox.MustNewAESCryptoV2(cfg.Resource.Key)).
			Register(cfg.Resource.Version, cryptox.MustNewAESCrypto(cfg.Resource.Key)).
			WithLegacyAlgo(cfg.User.Version),
		Runner: cryptox.NewCryptoManager("V2").
			Register("V2", cryptox.MustNewAESCryptoV2(cfg.Runner.Key)).
			Register(cfg.Runner.Version, cryptox.MustNewAESCrypto(cfg.Runner.Key)).
			WithLegacyAlgo(cfg.User.Version),
	}

	return reg
}
