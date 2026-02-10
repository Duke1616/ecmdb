//go:build wireinject

package order

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order/internal/event"
	"github.com/Duke1616/ecmdb/internal/order/internal/event/consumer"
	"github.com/Duke1616/ecmdb/internal/order/internal/grpc"
	"github.com/Duke1616/ecmdb/internal/order/internal/repository"
	"github.com/Duke1616/ecmdb/internal/order/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/internal/order/internal/web"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/sender"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	lark "github.com/larksuite/oapi-sdk-go/v3"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	service.NewProcessEngine,
	repository.NewOrderRepository,
	dao.NewTaskFormDAO,
	dao.NewOrderDAO,
	grpc.NewWorkOrderServer,
)

func InitModule(q mq.MQ, db *mongox.Mongo, workflowModule *workflow.Module, engineModule *engine.Module,
	templateModule *template.Module, userModule *user.Module, lark *lark.Client, sender sender.NotificationSender) (*Module, error) {
	wire.Build(
		ProviderSet,
		event.NewCreateProcessEventProducer,
		initWechatConsumer,
		InitProcessConsumer,
		InitModifyStatusConsumer,
		InitLardCallbackConsumer,
		wire.FieldsOf(new(*workflow.Module), "Svc"),
		wire.FieldsOf(new(*user.Module), "Svc"),
		wire.FieldsOf(new(*engine.Module), "Svc"),
		wire.FieldsOf(new(*template.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func initWechatConsumer(svc service.Service, templateSvc template.Service, userSvc user.Service, q mq.MQ) *consumer.WechatOrderConsumer {
	c, err := consumer.NewWechatOrderConsumer(svc, templateSvc, userSvc, q)
	if err != nil {
		panic(err)
	}

	c.Start(context.Background())
	return c
}

func InitProcessConsumer(q mq.MQ, workflowSvc workflow.Service, svc service.Service) *consumer.ProcessEventConsumer {
	c, err := consumer.NewProcessEventConsumer(q, workflowSvc, svc)
	if err != nil {
		return nil
	}

	c.Start(context.Background())
	return c
}

func InitModifyStatusConsumer(q mq.MQ, svc service.Service) *consumer.OrderStatusModifyEventConsumer {
	c, err := consumer.NewOrderStatusModifyEventConsumer(q, svc)
	if err != nil {
		return nil
	}

	c.Start(context.Background())
	return c
}

func InitLardCallbackConsumer(q mq.MQ, engineSvc engine.Service, lark *lark.Client, userSvc user.Service,
	templateSvc template.Service, svc service.Service, engineProcessSvc service.ProcessEngine,
	workflowSvc workflow.Service, sender sender.NotificationSender) *consumer.LarkCallbackEventConsumer {
	c, err := consumer.NewLarkCallbackEventConsumer(q, engineSvc, engineProcessSvc, svc, templateSvc,
		sender, userSvc, workflowSvc, lark)
	if err != nil {
		return nil
	}

	c.Start(context.Background())
	return c
}
