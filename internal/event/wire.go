//go:build wireinject

package event

import (
	easyEngine "github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/event/domain"
	"github.com/Duke1616/ecmdb/internal/event/producer"
	"github.com/Duke1616/ecmdb/internal/event/service/channel"
	"github.com/Duke1616/ecmdb/internal/event/service/easyflow"
	"github.com/Duke1616/ecmdb/internal/event/service/provider"
	"github.com/Duke1616/ecmdb/internal/event/service/provider/feishu"
	"github.com/Duke1616/ecmdb/internal/event/service/provider/sequential"
	"github.com/Duke1616/ecmdb/internal/event/service/sender"
	"github.com/Duke1616/ecmdb/internal/event/service/strategy"
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

var InitStrategySet = wire.NewSet(
	strategy.NewResult,
	strategy.NewUserNotification,
	strategy.NewAutomationNotification,
	strategy.NewStartNotification,
	strategy.NewDispatcher,
)

var InitSender = wire.NewSet(
	newSelectorBuilder,
	newChannel,
	sender.NewSender,
)

func InitModule(q mq.MQ, db *gorm.DB, engineModule *engine.Module, taskModule *task.Module, orderModule *order.Module,
	templateModule *template.Module, userModule *user.Module, workflowModule *workflow.Module,
	departmentModule *department.Module, lark *lark.Client) (*Module, error) {
	wire.Build(
		producer.NewOrderStatusModifyEventProducer,
		InitStrategySet,
		InitSender,
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

func InitWorkflowEngineOnce(db *gorm.DB, engineSvc engine.Service, producer producer.OrderStatusModifyEventProducer,
	taskSvc task.Service, orderSvc order.Service, workflowSvc workflow.Service,
	strategy strategy.SendStrategy) *easyflow.ProcessEvent {
	event, err := easyflow.NewProcessEvent(producer, engineSvc, taskSvc, orderSvc, workflowSvc, strategy)
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

func newChannel(builder *sequential.SelectorBuilder) channel.Channel {
	return channel.NewDispatcher(map[domain.Channel]channel.Channel{
		domain.ChannelFeishuCard: channel.NewFeishuCardChannel(builder),
	})
}

func newSelectorBuilder(
	lark *lark.Client,
) *sequential.SelectorBuilder {
	// 构建SMS供应商
	providers := make([]provider.Provider, 0)

	cardProvider, err := feishu.NewFeishuCardProvider(lark)
	if err != nil {
		return nil
	}

	providers = append(providers, cardProvider)
	return sequential.NewSelectorBuilder(providers)
}
