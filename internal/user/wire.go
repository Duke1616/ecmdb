//go:build wireinject

package user

import (
	"github.com/Duke1616/ecmdb/internal/user/internal/service"
	"github.com/Duke1616/ecmdb/internal/user/internal/web"
	"github.com/Duke1616/ecmdb/internal/user/ldapx"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	service.NewLdapService,
	web.NewHandler)

func InitModule(ldapConfig ldapx.Config) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
