// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package event

import (
	engine2 "github.com/Bunny3th/easy-workflow/workflow/engine"
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
	"github.com/larksuite/oapi-sdk-go/v3"
	"gorm.io/gorm"
	"log"
	"sync"
)

// Injectors from wire.go:

func InitModule(q mq.MQ, db *gorm.DB, engineModule *engine.Module, taskModule *task.Module, orderModule *order.Module, templateModule *template.Module, userModule *user.Module, workflowModule *workflow.Module, departmentModule *department.Module, lark2 *lark.Client) (*Module, error) {
	service := engineModule.Svc
	orderStatusModifyEventProducer, err := producer.NewOrderStatusModifyEventProducer(q)
	if err != nil {
		return nil, err
	}
	serviceService := taskModule.Svc
	service2 := orderModule.Svc
	service3 := workflowModule.Svc
	service4 := templateModule.Svc
	service5 := userModule.Svc
	service6 := departmentModule.Svc
	v := InitNotifyIntegration(lark2)
	v2 := InitNotification(service, service4, service2, service5, serviceService, service6, v)
	processEvent := InitWorkflowEngineOnce(db, service, orderStatusModifyEventProducer, serviceService, service2, service3, v2)
	module := &Module{
		Event: processEvent,
	}
	return module, nil
}

// wire.go:

var engineOnce = sync.Once{}

func InitNotifyIntegration(larkC *lark.Client) []method.NotifyIntegration {
	integrations, err := method.BuildReceiverIntegrations(larkC)
	if err != nil {

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

func InitWorkflowEngineOnce(db *gorm.DB, engineSvc engine.Service, producer2 producer.OrderStatusModifyEventProducer,
	taskSvc task.Service, orderSvc order.Service, workflowSvc workflow.Service,
	ns map[string]notification.SendAction) *easyflow.ProcessEvent {
	event, err := easyflow.NewProcessEvent(producer2, engineSvc, taskSvc, orderSvc, workflowSvc, ns)
	if err != nil {
		panic(err)
	}

	engineOnce.Do(func() {
		engine2.DB = db
		if err = engine2.DatabaseInitialize(); err != nil {
			log.Fatalln("easy workflow 初始化数据表失败，错误:", err)
		}
		engine2.IgnoreEventError = false
	})

	return event
}
