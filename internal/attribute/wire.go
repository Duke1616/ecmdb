//go:build wireinject

package attribute

import (
	"sync"

	"github.com/Duke1616/ecmdb/internal/attribute/internal/event"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/service"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	repository.NewAttributeRepository,
	repository.NewAttributeGroupRepository,
	dao.NewAttributeGroupDAO)

func InitModule(db *mongox.DB, q mq.MQ) (*Module, error) {
	wire.Build(
		ProviderSet,
		NewService,
		event.NewFieldSecureAttrChangeEventProducer,
		event.NewFieldDeleteEventProducer,
		InitAttributeDAO,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var daoOnce = sync.Once{}

func InitCollectionOnce(db *mongox.DB) {
	daoOnce.Do(func() {
		err := dao.InitIndexes(db)
		if err != nil {
			panic(err)
		}
	})
}

func InitAttributeDAO(db *mongox.DB) dao.AttributeDAO {
	InitCollectionOnce(db)
	return dao.NewAttributeDAO(db)
}

func NewService(repo repository.AttributeRepository, repoGroup repository.AttributeGroupRepository,
	producer event.FieldSecureAttrChangeEventProducer, deleteProducer event.IFieldDeleteEventProducer) Service {
	return service.NewService(repo, repoGroup, producer, deleteProducer)
}
