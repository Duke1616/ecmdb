package ioc

import (
	"github.com/Duke1616/etask/pkg/grpc/registry"
	"github.com/Duke1616/etask/pkg/grpc/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func InitRegistry(etcdClient *clientv3.Client) (registry.Registry, error) {
	return etcd.NewRegistry(etcdClient)
}
