//go:build wireinject

package event

import (
	easyEngine "github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/event/easyflow"
	"github.com/Duke1616/ecmdb/internal/event/easyflow/notification"
	"github.com/Duke1616/ecmdb/internal/event/easyflow/notification/method"
	"github.com/Duke1616/ecmdb/internal/event/easyflow/notification/node"
	"github.com/Duke1616/ecmdb/internal/event/easyflow/notification/result"
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
	departmentModule *department.Module, lark *lark.Client) (*Module, error) {
	wire.Build(
		producer.NewOrderStatusModifyEventProducer,
		InitNotifyIntegration,
		InitNotification,
		InitWorkflowEngineOnce,
		wire.FieldsOf(new(*engine.Module), "Svc"),
		wire.FieldsOf(new(*department.Module), "Svc"),
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

func InitNotifyIntegration(larkC *lark.Client) []method.NotifyIntegration {
	integrations, err := method.BuildReceiverIntegrations(larkC)
	if err != nil {
		// TODO 记录日志
		return nil
	}

	return integrations
}

func InitNotification(engineSvc engine.Service, templateSvc template.Service, orderSvc order.Service,
	userSvc user.Service, taskSvc task.Service, departMentSvc department.Service,
	integration []method.NotifyIntegration) map[string]notification.SendAction {
	resultSvc := result.NewResult(taskSvc)
	
	ns := make(map[string]notification.SendAction)
	userNotify, err := node.NewUserNotification(engineSvc, templateSvc, orderSvc, userSvc, resultSvc,
		departMentSvc, integration)
	if err != nil {
		panic(err)
	}

	automationNotify, err := node.NewAutomationNotification(resultSvc, userSvc, integration)
	if err != nil {
		panic(err)
	}

	startNotify, err := node.NewStartNotification(userSvc, templateSvc, integration)
	if err != nil {
		panic(err)
	}

	ns["user"] = userNotify
	ns["automation"] = automationNotify
	ns["start"] = startNotify
	return ns
}

func InitWorkflowEngineOnce(db *gorm.DB, engineSvc engine.Service, producer producer.OrderStatusModifyEventProducer,
	taskSvc task.Service, orderSvc order.Service, workflowSvc workflow.Service,
	ns map[string]notification.SendAction) *easyflow.ProcessEvent {
	event, err := easyflow.NewProcessEvent(producer, engineSvc, taskSvc, orderSvc, workflowSvc, ns)
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
