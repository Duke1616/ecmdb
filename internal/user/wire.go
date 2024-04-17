//go:build wireinject

package user

import (
	"github.com/Duke1616/ecmdb/internal/user/internal/repostory"
	"github.com/Duke1616/ecmdb/internal/user/internal/repostory/dao"
	"github.com/Duke1616/ecmdb/internal/user/internal/service"
	"github.com/Duke1616/ecmdb/internal/user/internal/web"
	"github.com/Duke1616/ecmdb/internal/user/ldapx"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
)

var ProviderSet = wire.NewSet(
	service.NewLdapService,
	service.NewService,
	repostory.NewResourceRepository,
	dao.NewUserDao,
	web.NewHandler)

func InitModule(db *mongo.Client, ldapConfig ldapx.Config) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
