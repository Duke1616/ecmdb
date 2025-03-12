//go:build wireinject

package task

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/runner"
	"github.com/Duke1616/ecmdb/internal/task/internal/event"
	"github.com/Duke1616/ecmdb/internal/task/internal/job"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/Duke1616/ecmdb/internal/task/internal/web"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	"time"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewExecService,
	service.NewService,
	repository.NewTaskRepository,
	dao.NewTaskDAO,
)

func InitModule(q mq.MQ, db *mongox.Mongo, orderModule *order.Module, workflowModule *workflow.Module,
	engineModule *engine.Module, codebookModule *codebook.Module, workerModule *worker.Module,
	runnerModule *runner.Module, userModule *user.Module, lark *lark.Client) (*Module, error) {
	wire.Build(
		ProviderSet,
		initStartTaskJob,
		initPassProcessTaskJob,
		initConsumer,
		wire.FieldsOf(new(*order.Module), "Svc"),
		wire.FieldsOf(new(*workflow.Module), "Svc"),
		wire.FieldsOf(new(*codebook.Module), "Svc"),
		wire.FieldsOf(new(*user.Module), "Svc"),
		wire.FieldsOf(new(*worker.Module), "Svc"),
		wire.FieldsOf(new(*runner.Module), "Svc"),
		wire.FieldsOf(new(*engine.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func initConsumer(svc service.Service, q mq.MQ, codebookSvc codebook.Service,
	userSvc user.Service, lark *lark.Client) *event.ExecuteResultConsumer {
	consumer, err := event.NewExecuteResultConsumer(q, svc, codebookSvc, userSvc, lark)
	if err != nil {
		panic(err)
	}

	consumer.Start(context.Background())
	return consumer
}

func initStartTaskJob(svc service.Service) *StartTaskJob {
	limit := int64(100)
	initialInterval := 10 * time.Second
	maxInterval := 30 * time.Second
	maxRetries := int32(3)
	return job.NewStartTaskJob(svc, limit, initialInterval, maxInterval, maxRetries)
}

func initPassProcessTaskJob(svc service.Service, engineSvc engine.Service) *PassProcessTaskJob {
	minutes := int64(30)
	seconds := int64(10)
	limit := int64(100)
	return job.NewPassProcessTaskJob(svc, engineSvc, minutes, seconds, limit)
}
