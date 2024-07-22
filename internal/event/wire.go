package event

import (
	easyEngine "github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/event/easyflow"
	"github.com/Duke1616/ecmdb/internal/event/producer"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	"gorm.io/gorm"
	"log"
	"sync"
)

func InitModule(q mq.MQ, db *gorm.DB, engineModule *engine.Module) (*Module, error) {
	wire.Build(
		producer.NewOrderStatusModifyEventProducer,
		InitWorkflowEngineOnce,
		wire.FieldsOf(new(*engine.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var engineOnce = sync.Once{}

func InitWorkflowEngineOnce(db *gorm.DB, svc engine.Service, producer producer.OrderStatusModifyEventProducer,
) *easyflow.ProcessEvent {
	event := easyflow.NewProcessEvent(svc, producer)
	engineOnce.Do(func() {
		easyEngine.DB = db
		if err := easyEngine.DatabaseInitialize(); err != nil {
			log.Fatalln("easy workflow 初始化数据表失败，错误:", err)
		}
		// 是否忽略事件错误
		easyEngine.IgnoreEventError = false
	})

	return event
}
