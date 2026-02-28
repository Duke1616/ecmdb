//go:build wireinject

package template

import (
	"context"
	"sync"

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
	InitFavoriteDAO,
	web.NewGroupHandler,
	service.NewGroupService,
	repository.NewTemplateGroupRepository,
	dao.NewTemplateGroupDAO,
)

var daoOnce = sync.Once{}

func InitCollectionOnce(db *mongox.Mongo) {
	daoOnce.Do(func() {
		err := dao.InitIndexes(db)
		if err != nil {
			panic(err)
		}
	})
}

func InitFavoriteDAO(db *mongox.Mongo) dao.FavoriteDAO {
	InitCollectionOnce(db)
	return dao.NewFavoriteDAO(db)
}

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
