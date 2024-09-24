//go:build wireinject

package event

import (
	easyEngine "github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/event/easyflow"
	"github.com/Duke1616/ecmdb/internal/event/producer"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	"gorm.io/gorm"
	"log"
	"sync"
)

func InitModule(q mq.MQ, db *gorm.DB, engineModule *engine.Module, taskModule *task.Module, orderModule *order.Module,
	templateModule *template.Module, userModule *user.Module, workflowModule *workflow.Module,
	lark *lark.Client) (*Module, error) {
	wire.Build(
		producer.NewOrderStatusModifyEventProducer,
		InitNotifyIntegration,
		InitWorkflowEngineOnce,
		wire.FieldsOf(new(*engine.Module), "Svc"),
		wire.FieldsOf(new(*task.Module), "Svc"),
		wire.FieldsOf(new(*template.Module), "Svc"),
		wire.FieldsOf(new(*order.Module), "Svc"),
		wire.FieldsOf(new(*workflow.Module), "Svc"),
		wire.FieldsOf(new(*user.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var engineOnce = sync.Once{}

func InitNotifyIntegration(larkC *lark.Client) []easyflow.NotifyIntegration {
	integrations, err := easyflow.BuildReceiverIntegrations(larkC)
	if err != nil {
		// TODO 记录日志
		return nil
	}

	return integrations
}

func InitWorkflowEngineOnce(db *gorm.DB, engineSvc engine.Service, producer producer.OrderStatusModifyEventProducer,
	taskSvc task.Service, templateSvc template.Service, orderSvc order.Service, userSvc user.Service,
	workflowSvc workflow.Service, integration []easyflow.NotifyIntegration) *easyflow.ProcessEvent {
	// 消息通知
	notify, err := easyflow.NewNotify(engineSvc, templateSvc, orderSvc, userSvc, workflowSvc, integration)
	if err != nil {
		panic(err)
	}

	// 注册event
	event, err := easyflow.NewProcessEvent(producer, engineSvc, taskSvc, notify, orderSvc)
	if err != nil {
		panic(err)
	}

	engineOnce.Do(func() {
		easyEngine.DB = db
		if err = easyEngine.DatabaseInitialize(); err != nil {
			log.Fatalln("easy workflow 初始化数据表失败，错误:", err)
		}
		// 是否忽略事件错误
		easyEngine.IgnoreEventError = false
	})

	return event
}
