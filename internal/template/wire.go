//go:build wireinject

package template

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/template/internal/event"
	"github.com/Duke1616/ecmdb/internal/template/internal/repository"
	"github.com/Duke1616/ecmdb/internal/template/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/template/internal/service"
	"github.com/Duke1616/ecmdb/internal/template/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	"github.com/xen0n/go-workwx"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewTemplateRepository,
	dao.NewTemplateDAO,
	web.NewGroupHandler,
	service.NewGroupService,
	repository.NewTemplateGroupRepository,
	dao.NewTemplateGroupDAO,
)

func InitModule(q mq.MQ, db *mongox.Mongo, workAPP *workwx.WorkwxApp) (*Module, error) {
	wire.Build(
		ProviderSet,
		event.NewWechatOrderEventProducer,
		initConsumer,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func initConsumer(svc service.Service, q mq.MQ, p event.WechatOrderEventProducer, workAPP *workwx.WorkwxApp) *event.WechatApprovalCallbackConsumer {
	consumer, err := event.NewWechatApprovalCallbackConsumer(svc, q, p, workAPP)
	if err != nil {
		panic(err)
	}

	consumer.Start(context.Background())
	return consumer
}
