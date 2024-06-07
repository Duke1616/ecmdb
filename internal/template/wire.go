//go:build wireinject

package template

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/template/internal/event"
	"github.com/Duke1616/ecmdb/internal/template/internal/service"
	"github.com/Duke1616/ecmdb/internal/template/internal/web"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	"github.com/xen0n/go-workwx"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService)

func InitModule(q mq.MQ, workAPP *workwx.WorkwxApp) (*Module, error) {
	wire.Build(
		ProviderSet,
		initConsumer,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func initConsumer(svc service.Service, q mq.MQ, workAPP *workwx.WorkwxApp) *event.WechatApprovalCallbackConsumer {
	consumer, err := event.NewWechatApprovalCallbackConsumer(svc, q, workAPP)
	if err != nil {
		panic(err)
	}

	consumer.Start(context.Background())
	return consumer
}
