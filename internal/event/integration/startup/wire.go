//go:build wireinject

package startup

import (
	notificationv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/notification/v1"
	teamv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/team"
	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/event"
	"github.com/Duke1616/ecmdb/internal/event/producer"
	"github.com/Duke1616/ecmdb/internal/event/service/easyflow"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/sender"
	"github.com/Duke1616/ecmdb/internal/rota"
	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/test/ioc"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/google/wire"
	lark "github.com/larksuite/oapi-sdk-go/v3"
)

type TestApp struct {
	EventHandler     *easyflow.ProcessEvent
	UserModule       *user.Module
	OrderModule      *order.Module
	TaskModule       *task.Module
	TemplateModule   *template.Module
	EngineModule     *engine.Module
	WorkflowModule   *workflow.Module
	DepartmentModule *department.Module
	RotaModule       *rota.Module
}

func InitApp(
	sender sender.NotificationSender,
	teamSvc teamv1.TeamServiceClient,
	notificationSvc notificationv1.NotificationServiceClient,
	userModule *user.Module,
	orderModule *order.Module,
	taskModule *task.Module,
	templateModule *template.Module,
	engineModule *engine.Module,
	workflowModule *workflow.Module,
	departmentModule *department.Module,
	rotaModule *rota.Module,
) (*TestApp, error) {
	wire.Build(
		ioc.BaseSet,

		// External Clients (Nil for integration tests)
		wire.Value((*lark.Client)(nil)),

		// Event Module components (The core logic we are testing)
		event.InitStrategySet,
		event.InitWorkflowEngineOnce,
		producer.NewOrderStatusModifyEventProducer,

		// Extract Services from provided Modules
		wire.FieldsOf(new(*user.Module), "Svc"),
		wire.FieldsOf(new(*order.Module), "Svc"),
		wire.FieldsOf(new(*task.Module), "Svc"),
		wire.FieldsOf(new(*template.Module), "Svc"),
		wire.FieldsOf(new(*engine.Module), "Svc"),
		wire.FieldsOf(new(*workflow.Module), "Svc"),
		wire.FieldsOf(new(*department.Module), "Svc"),
		wire.FieldsOf(new(*rota.Module), "Svc"),

		wire.Struct(new(TestApp), "*"),
	)
	return nil, nil
}
