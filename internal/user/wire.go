//go:build wireinject

package user

import (
	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/user/internal/repository"
	"github.com/Duke1616/ecmdb/internal/user/internal/repository/cache"
	"github.com/Duke1616/ecmdb/internal/user/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/user/internal/service"
	"github.com/Duke1616/ecmdb/internal/user/internal/web"
	"github.com/Duke1616/ecmdb/internal/user/ldapx"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

var ProviderSet = wire.NewSet(
	service.NewLdapService,
	service.NewService,
	repository.NewResourceRepository,
	dao.NewUserDao,
	web.NewHandler,
)

func InitLdapUserCache(redisClient redis.Cmdable) cache.LdapUserCache {
	return cache.NewLdapUserCache(redisClient, 0)
}

func InitModule(db *mongox.Mongo, redisClient redis.Cmdable, ldapConfig ldapx.Config, policyModule *policy.Module,
	departmentModule *department.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitLdapUserCache,
		wire.Struct(new(Module), "*"),
		wire.FieldsOf(new(*department.Module), "Svc"),
		wire.FieldsOf(new(*policy.Module), "Svc"),
	)
	return new(Module), nil
}
