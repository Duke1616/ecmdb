//go:build wireinject

package resource

import (
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/resource/internal/service"
	"github.com/Duke1616/ecmdb/internal/resource/internal/web"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewResourceRepository,
	dao.NewResourceDAO)

func InitHandler(db *mongo.Client) *Handler {
	wire.Build(ProviderSet)
	return new(Handler)
}

type Handler = web.Handler
