package ioc

import (
	"fmt"

	"github.com/RediSearch/redisearch-go/v2/redisearch"
	"github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
)

func InitRedisSearch() *redisearch.Client {
	type Config struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("redis", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into structure: %v", err))
	}

	pool := &redis.Pool{Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", cfg.Addr,
			redis.DialPassword(cfg.Password),
			redis.DialDatabase(cfg.DB))
	}}

	return redisearch.NewClientFromPool(pool, "index")
}
