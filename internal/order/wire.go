//go:build wireinject

package order

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/order/internal/event"
	"github.com/Duke1616/ecmdb/internal/order/internal/repository"
	"github.com/Duke1616/ecmdb/internal/order/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/internal/order/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewOrderRepository,
	dao.NewOrderDAO,
)

func InitModule(q mq.MQ, db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		initConsumer,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func initConsumer(svc service.Service, q mq.MQ) *event.WechatOrderConsumer {
	consumer, err := event.NewWechatOrderConsumer(svc, q)
	if err != nil {
		panic(err)
	}

	consumer.Start(context.Background())
	return consumer
}
