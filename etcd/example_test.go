package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
)

func TestPutEtcd(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"10.31.0.200:2379"},
	})
	if err != nil {
		panic(err)
	}

	put, err := etcdClient.Put(context.Background(), "/worker/name", "123")
	if err != nil {
		return
	}

	fmt.Println(put)
}
