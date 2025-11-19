//go:build wireinject

package ioc

import (
	"github.com/Duke1616/ecmdb/cmd/initial/version"
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/bootstrap"
	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/menu"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/permission"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/internal/role"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/ioc"
	"github.com/google/wire"
)

var BaseSet = wire.NewSet(ioc.InitMongoDB, ioc.InitMySQLDB, ioc.InitRedis, ioc.InitRedisSearch,
	ioc.InitMQ, ioc.InitEtcdClient, ioc.InitLdapConfig, ioc.InitModuleCrypto)

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
		// 关联关系模块（无依赖，最先初始化）
		relation.InitModule,
		// 字段模块（依赖 MQ）
		attribute.InitModule,
		// 资源模块（model 依赖它）
		resource.InitModule,
		// 模型模块（依赖 relation、attribute、resource）
		model.InitModule,
		// Bootstrap 模块（依赖 model、attribute、relation，统一管理 CMDB 相关服务）
		bootstrap.InitModule,
		wire.FieldsOf(new(*bootstrap.Module), "Svc"),
	)
	return new(App), nil
}
