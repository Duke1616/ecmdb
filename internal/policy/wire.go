//go:build wireinject

package policy

import (
	"github.com/Duke1616/ecmdb/internal/policy/internal/grpc"
	"github.com/Duke1616/ecmdb/internal/policy/internal/service"
	"github.com/Duke1616/ecmdb/internal/policy/internal/web"
	"github.com/casbin/casbin/v2"
	"github.com/ecodeclub/ginx/session"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	grpc.NewPolicyServer,
)

func InitModule(enforcer *casbin.SyncedEnforcer, sp session.Provider) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
