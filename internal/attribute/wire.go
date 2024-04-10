//go:build wireinject

package attribute

import (
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/service"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/web"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewAttributeRepository,
	dao.NewAttributeDAO)

func InitHandler(db *mongo.Client) *Handler {
	wire.Build(ProviderSet)
	return new(Handler)
}

type Handler = web.Handler
