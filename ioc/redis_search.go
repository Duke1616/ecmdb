package ioc

import (
	"github.com/RediSearch/redisearch-go/v2/redisearch"
	"github.com/gomodule/redigo/redis"
)

func InitRedisSearch() *redisearch.Client {
	type Config struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	}

	var cfg Config
	pool := &redis.Pool{Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", cfg.Addr,
			redis.DialPassword(cfg.Password),
			redis.DialDatabase(cfg.DB))
	}}

	return redisearch.NewClientFromPool(pool, "index")
}
