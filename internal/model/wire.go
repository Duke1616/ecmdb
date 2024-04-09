//go:build wireinject

package model

import (
	"github.com/Duke1616/ecmdb/internal/model/internal/repository"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/model/internal/service"
	"github.com/Duke1616/ecmdb/internal/model/internal/web"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewModelRepository,
	dao.NewModelDAO)

func InitHandler(db *mongo.Client) *Handler {
	wire.Build(ProviderSet)
	return new(Handler)
}

type Handler = web.Handler
