package startup

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func InitEtcdClient() *clientv3.Client {
	var cfg clientv3.Config

	// Unmarshal etcd configuration from viper
	if err := viper.UnmarshalKey("etcd", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into structure: %v", err))
	}

	// Set connection timeout (e.g., 5 seconds)
	cfg.DialTimeout = 5 * time.Second

	// Create the etcd client
	client, err := clientv3.New(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to connect to etcd: %v", err))
	}

	// Perform a ping test to ensure the etcd server is reachable
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get the status of the first available endpoint
	_, err = client.Status(ctx, client.Endpoints()[0])
	if err != nil {
		panic(fmt.Errorf("failed to ping etcd: %v", err))
	}

	return client
}
