//go:build wireinject

package ioc

import (
	"github.com/Duke1616/ecmdb/cmd/initial/version"
	"github.com/Duke1616/ecmdb/ioc"
	"github.com/google/wire"
)

var BaseSet = wire.NewSet(
	ioc.InitMongoDB,
	ioc.InitMongoDBV2,
	ioc.InitRedis,
	ioc.InitMQ,
	ioc.InitEtcdClient,
	ioc.InitCrypto,
)

func InitApp() (*App, error) {
	wire.Build(
		wire.Struct(new(App), "*"),
		BaseSet,
		version.NewService,
		version.NewDao,
	)
	return new(App), nil
}
