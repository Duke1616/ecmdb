package ioc

import (
	"fmt"
	"github.com/ecodeclub/ekit/retry"
	"github.com/ecodeclub/mq-api"
	"github.com/ecodeclub/mq-api/kafka"
	"github.com/spf13/viper"
	"sync"
	"time"
)

var (
	q          mq.MQ
	mqInitOnce sync.Once
)

func InitMQ(viper *viper.Viper) mq.MQ {
	mqInitOnce.Do(func() {
		const maxInterval = 10 * time.Second
		const maxRetries = 10
		strategy, err := retry.NewExponentialBackoffRetryStrategy(time.Second, maxInterval, maxRetries)
		if err != nil {
			panic(err)
		}
		for {
			q, err = initMQ(viper)
			if err == nil {
				break
			}
			next, ok := strategy.Next()
			if !ok {
				panic("InitMQ 重试失败......")
			}
			time.Sleep(next)
		}
	})
	return q
}

func initMQ(viper *viper.Viper) (mq.MQ, error) {
	type Config struct {
		Network   string   `yaml:"network"`
		Addresses []string `yaml:"addresses"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("kafka", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %v", err))
	}

	qq, err := kafka.NewMQ(cfg.Network, cfg.Addresses)
	if err != nil {
		return nil, err
	}
	return qq, nil
}
