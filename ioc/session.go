package ioc

import (
	"fmt"

	"github.com/ecodeclub/ginx/session"
	ginRedis "github.com/ecodeclub/ginx/session/redis"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitSession(cmd redis.Cmdable) session.Provider {
	type Config struct {
		SessionEncryptedKey string `mapstructure:"session_encrypted_key"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("session", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %v", err))
	}

	sp := ginRedis.NewSessionProvider(cmd, cfg.SessionEncryptedKey)

	return sp
}
