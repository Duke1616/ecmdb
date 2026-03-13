//go:build wireinject

package event

import (
	"log"
	"sync"

	easyEngine "github.com/Bunny3th/easy-workflow/workflow/engine"
	notificationv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/notification/v1"
	teamv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/team"
	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/event/producer"
	"github.com/Duke1616/ecmdb/internal/event/service/assignees"
	"github.com/Duke1616/ecmdb/internal/event/service/easyflow"
	"github.com/Duke1616/ecmdb/internal/event/service/strategy"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/sender"
	"github.com/Duke1616/ecmdb/internal/rota"
	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/pkg/resolve"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	"gorm.io/gorm"
)

var InitStrategySet = wire.NewSet(
	strategy.NewService,
	strategy.NewUserNotification,
	strategy.NewAutomationNotification,
	strategy.NewStartNotification,
	strategy.NewChatNotification,
	strategy.NewCarbonCopyNotification,
	strategy.NewDispatcher,
	wire.Bind(new(strategy.SendStrategy), new(*strategy.Dispatcher)),

	// Resolvers
	assignees.NewAppointResolver,
	assignees.NewFounderResolver,
	assignees.NewLeaderResolver,
	assignees.NewMainLeaderResolver,
	assignees.NewOnCallResolver,
	assignees.NewTemplateResolver,
	assignees.NewTeamResolver,
	InitResolveEngine,
)

func InitResolveEngine(
	appoint *assignees.AppointResolver,
	founder *assignees.FounderResolver,
	leader *assignees.LeaderResolver,
	mainLeader *assignees.MainLeaderResolver,
	onCall *assignees.OnCallResolver,
	template *assignees.TemplateResolver,
	team *assignees.TeamResolver,
) *resolve.Engine {
	return resolve.NewEngine().Register(
		appoint,
		founder,
		leader,
		mainLeader,
		onCall,
		template,
		team,
	)
}

func InitModule(q mq.MQ, db *gorm.DB, engineModule *engine.Module, taskModule *task.Module, orderModule *order.Module,
	templateModule *template.Module, userModule *user.Module, workflowModule *workflow.Module, sender sender.NotificationSender,
	departmentModule *department.Module, rotaModule *rota.Module, lark *lark.Client, notificationSvc notificationv1.NotificationServiceClient,
	teamSvc teamv1.TeamServiceClient) (*Module, error) {
	wire.Build(
		producer.NewOrderStatusModifyEventProducer,
		InitStrategySet,
		InitWorkflowEngineOnce,
		wire.FieldsOf(new(*engine.Module), "Svc"),
		wire.FieldsOf(new(*department.Module), "Svc"),
		wire.FieldsOf(new(*task.Module), "Svc"),
		wire.FieldsOf(new(*template.Module), "Svc"),
		wire.FieldsOf(new(*order.Module), "Svc"),
		wire.FieldsOf(new(*workflow.Module), "Svc"),
		wire.FieldsOf(new(*user.Module), "Svc"),
		wire.FieldsOf(new(*rota.Module), "Svc"),
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
