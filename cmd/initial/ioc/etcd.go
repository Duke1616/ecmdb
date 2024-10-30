package ioc

import (
	"fmt"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func InitEtcdClient() *clientv3.Client {
	var cfg clientv3.Config

	if err := viper.UnmarshalKey("etcd", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %v", err))
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		panic(err)
	}

	return client
}
