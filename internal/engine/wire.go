//go:build wireinject

package engine

import (
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Duke1616/ecmdb/internal/engine/event"
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
)

func InitModule(db *gorm.DB, q mq.MQ) (*Module, error) {
	wire.Build(
		ProviderSet,
		event.NewProcessEvent,
		event.NewOrderStatusModifyEventProducer,
		InitWorkflowEngineOnce,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var engineOnce = sync.Once{}

func InitWorkflowEngineOnce(db *gorm.DB, event *event.ProcessEvent) dao.ProcessEngineDAO {
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

	return dao.NewProcessEngineDAO(db)
}
