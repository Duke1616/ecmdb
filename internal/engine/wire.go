//go:build wireinject

package engine

import (
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Duke1616/ecmdb/internal/engine/internal/repository"
	"github.com/Duke1616/ecmdb/internal/engine/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/engine/internal/service"
	"github.com/Duke1616/ecmdb/internal/engine/internal/web"
	"github.com/Duke1616/ecmdb/internal/engine/register"
	"github.com/google/wire"
	"gorm.io/gorm"
	"log"
	"sync"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewTaskRepository,
)

func InitModule(db *gorm.DB) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitEasyFlowOnce,
		//InitConsumer,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var flowOnce = sync.Once{}

func InitEasyFlowOnce(db *gorm.DB) dao.ProcessEngineDAO {
	flowOnce.Do(func() {
		engine.DB = db
		if err := engine.DatabaseInitialize(); err != nil {
			log.Fatalln("easy workflow 初始化数据表失败，错误:", err)
		}
		//是否忽略事件错误
		engine.IgnoreEventError = false

		//注册事件函数
		engine.RegisterEvents(&register.EasyFlowEvent{})

		// 服务启动成功
		log.Println("========================== easy workflow 启动成功 ========================== ")
	})

	return dao.NewProcessEngineDAO(db)
}

//func InitConsumer(q mq.MQ, workflowSvc workflow.Service, orderSvc order.Service) *event.ProcessEventConsumer {
//	consumer, err := event.NewProcessEventConsumer(q, workflowSvc, orderSvc)
//	if err != nil {
//		return nil
//	}
//
//	consumer.Start(context.Background())
//	return consumer
//}
