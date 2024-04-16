package user

import (
	"github.com/Duke1616/ecmdb/internal/user/internal/repostory"
	"github.com/Duke1616/ecmdb/internal/user/internal/repostory/dao"
	"github.com/Duke1616/ecmdb/internal/user/internal/service"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
)

var ProviderSet = wire.NewSet(
	service.NewService,
	repostory.NewResourceRepository,
	dao.NewUserDao)

func InitModule(db *mongo.Client) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
