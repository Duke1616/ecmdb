//go:build wireinject

package relation

import (
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/internal/relation/internal/web"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewRelationRepository,
	dao.NewRelationDAO)

func InitHandler(db *mongo.Client) *Handler {
	wire.Build(ProviderSet)
	return new(Handler)
}

type Handler = web.Handler
