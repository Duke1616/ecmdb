//go:build wireinject

package task

import (
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Duke1616/ecmdb/internal/task/register"
	"github.com/Duke1616/ecmdb/internal/task/web"
	"github.com/google/wire"
	"gorm.io/gorm"
	"log"
	"sync"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
)

func InitModule(db *gorm.DB) (*Module, error) {
	wire.Build(
		//ProviderSet,
		InitEasyFlowOnce,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var flowOnce = sync.Once{}

func InitEasyFlowOnce(db *gorm.DB) *web.Handler {
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

	return web.NewHandler()
}
