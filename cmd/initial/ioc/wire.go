//go:build wireinject

package ioc

import (
	"github.com/Duke1616/ecmdb/cmd/initial/version"
	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/role"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/google/wire"
)

var BaseSet = wire.NewSet(InitMongoDB, InitMySQLDB, InitRedis, InitRediSearch, InitMQ, InitEtcdClient, InitLdapConfig)

func InitApp() (*App, error) {
	wire.Build(wire.Struct(new(App), "*"),
		BaseSet,
		InitCasbin,
		InitSession,
		user.InitModule,
		version.NewService,
		version.NewDao,
		wire.FieldsOf(new(*user.Module), "Svc"),
		department.InitModule,
		role.InitModule,
		wire.FieldsOf(new(*role.Module), "Svc"),
		policy.InitModule,
	)
	return new(App), nil
}
