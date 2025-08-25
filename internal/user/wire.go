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
	"github.com/RediSearch/redisearch-go/v2/redisearch"
	"github.com/ecodeclub/ginx/session"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	service.NewLdapService,
	service.NewService,
	repository.NewResourceRepository,
	dao.NewUserDao,
	web.NewHandler,
)

func InitLdapUserCache(conn *redisearch.Client) cache.RedisearchLdapUserCache {
	return cache.NewRedisearchLdapUserCache(conn)
}

func InitModule(db *mongox.Mongo, redisClient *redisearch.Client, ldapConfig ldapx.Config, policyModule *policy.Module,
	departmentModule *department.Module, sp session.Provider) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitLdapUserCache,
		wire.Struct(new(Module), "*"),
		wire.FieldsOf(new(*department.Module), "Svc"),
		wire.FieldsOf(new(*policy.Module), "Svc"),
	)
	return new(Module), nil
}
