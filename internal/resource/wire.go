//go:build wireinject

package resource

import (
	"context"
	"sync"

	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource/internal/event"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/resource/internal/service"
	"github.com/Duke1616/ecmdb/internal/resource/internal/web"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	repository.NewResourceRepository)

func InitModule(db *mongox.Mongo, attributeModule *attribute.Module, relationModule *relation.Module,
	q mq.MQ, crypto *cryptox.CryptoRegistry) (*Module, error) {
	wire.Build(
		ProviderSet,
		NewEncryptedService,
		InitResourceDAO,
		NewService,
		InitCrypto,
		initConsumer,
		wire.FieldsOf(new(*attribute.Module), "Svc"),
		wire.FieldsOf(new(*relation.Module), "RRSvc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var daoOnce = sync.Once{}

func InitCollectionOnce(db *mongox.Mongo) {
	daoOnce.Do(func() {
		err := dao.InitIndexes(db)
		if err != nil {
			panic(err)
		}
	})
}

func InitResourceDAO(db *mongox.Mongo) dao.ResourceDAO {
	InitCollectionOnce(db)
	return dao.NewResourceDAO(db)
}

func NewService(repo repository.ResourceRepository) Service {
	return service.NewService(repo)
}

func NewEncryptedService(baseSvc service.Service, attrSvc attribute.Service,
	cryptox cryptox.Crypto[string]) EncryptedSvc {
	return service.NewEncryptedResourceService(baseSvc, attrSvc, cryptox)
}

func InitCrypto(reg *cryptox.CryptoRegistry) cryptox.Crypto[string] {
	return reg.Resource
}

func initConsumer(q mq.MQ, svc service.EncryptedSvc, cryptox cryptox.Crypto[string]) *event.FieldSecureAttrChangeConsumer {
	consumer, err := event.NewFieldSecureAttrChangeConsumer(q, svc, 20, cryptox)
	if err != nil {
		panic(err)
	}

	consumer.Start(context.Background())
	return consumer
}
