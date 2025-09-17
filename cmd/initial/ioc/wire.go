//go:build wireinject

package ioc

import (
	"github.com/Duke1616/ecmdb/cmd/initial/version"
	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/Duke1616/ecmdb/internal/permission"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/role"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/ioc"
	"github.com/google/wire"
)

var BaseSet = wire.NewSet(ioc.InitMongoDB, ioc.InitMySQLDB, ioc.InitRedis, ioc.InitRediSearch,
	ioc.InitMQ, ioc.InitEtcdClient, ioc.InitLdapConfig, ioc.AesKey)

func InitApp() (*App, error) {
	wire.Build(wire.Struct(new(App), "*"),
		BaseSet,
		ioc.InitCasbin,
		ioc.InitSession,
		user.InitModule,
		version.NewService,
		version.NewDao,
		wire.FieldsOf(new(*user.Module), "Svc"),
		department.InitModule,
		role.InitModule,
		wire.FieldsOf(new(*role.Module), "Svc"),
		menu.InitModule,
		wire.FieldsOf(new(*menu.Module), "Svc"),
		permission.InitModule,
		wire.FieldsOf(new(*permission.Module), "Svc"),
		policy.InitModule,
		wire.FieldsOf(new(*policy.Module), "Svc"),
	)
	return new(App), nil
}
