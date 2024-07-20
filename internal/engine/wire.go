//go:build wireinject

package engine

import (
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	engineEvent "github.com/Duke1616/ecmdb/internal/engine/internal/event/engine"
	"github.com/Duke1616/ecmdb/internal/engine/internal/event/producer"
	"github.com/Duke1616/ecmdb/internal/engine/internal/repository"
	"github.com/Duke1616/ecmdb/internal/engine/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/engine/internal/service"
	"github.com/Duke1616/ecmdb/internal/engine/internal/web"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	"gorm.io/gorm"
	"log"
	"sync"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewProcessEngineRepository,
	dao.NewProcessEngineDAO,
)

func InitModule(db *gorm.DB, q mq.MQ) (*Module, error) {
	wire.Build(
		ProviderSet,
		producer.NewOrderStatusModifyEventProducer,
		InitWorkflowEngineOnce,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var engineOnce = sync.Once{}

func InitWorkflowEngineOnce(db *gorm.DB, producer producer.OrderStatusModifyEventProducer, svc service.Service) *engineEvent.ProcessEvent {
	event := engineEvent.NewProcessEvent(producer, svc)

	engineOnce.Do(func() {
		engine.DB = db
		if err := engine.DatabaseInitialize(); err != nil {
			log.Fatalln("easy workflow 初始化数据表失败，错误:", err)
		}
		// 是否忽略事件错误
		engine.IgnoreEventError = false

		// 事件注册
		engine.RegisterEvents(event)
	})

	return event
}
