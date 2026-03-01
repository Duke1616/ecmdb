//go:build wireinject

package task

import (
	"context"
	"sync"
	"time"

	executorv1 "github.com/Duke1616/ecmdb/api/proto/gen/etask/executor/v1"
	taskv1 "github.com/Duke1616/ecmdb/api/proto/gen/etask/task/v1"
	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/discovery"
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/sender"
	"github.com/Duke1616/ecmdb/internal/runner"
	"github.com/Duke1616/ecmdb/internal/task/internal/event"
	"github.com/Duke1616/ecmdb/internal/task/internal/job"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/Duke1616/ecmdb/internal/task/internal/service/dispatch"
	"github.com/Duke1616/ecmdb/internal/task/internal/service/scheduler"
	"github.com/Duke1616/ecmdb/internal/task/internal/web"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	lark "github.com/larksuite/oapi-sdk-go/v3"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	dispatch.NewTaskDispatcher,
	service.NewService,
	repository.NewTaskRepository,
	InitTaskDAO,
	scheduler.NewScheduler,
)

func InitModule(q mq.MQ, db *mongox.Mongo, orderModule *order.Module, workflowModule *workflow.Module,
	engineModule *engine.Module, codebookModule *codebook.Module, workerModule *worker.Module,
	runnerModule *runner.Module, userModule *user.Module, discoveryModule *discovery.Module,
	lark *lark.Client, crypto *cryptox.CryptoRegistry, sender sender.NotificationSender,
	taskClient taskv1.TaskServiceClient, executorClient executorv1.TaskExecutionServiceClient) (*Module, error) {
	wire.Build(
		ProviderSet,
		initStartTaskJob,
		initPassProcessTaskJob,
		initTaskRecoveryJob,
		initTaskExecutionSyncJob,
		initConsumer,
		InitCrypto,
		wire.FieldsOf(new(*order.Module), "Svc"),
		wire.FieldsOf(new(*workflow.Module), "Svc"),
		wire.FieldsOf(new(*codebook.Module), "Svc"),
		wire.FieldsOf(new(*user.Module), "Svc"),
		wire.FieldsOf(new(*worker.Module), "Svc"),
		wire.FieldsOf(new(*discovery.Module), "Svc"),
		wire.FieldsOf(new(*runner.Module), "Svc"),
		wire.FieldsOf(new(*engine.Module), "Svc"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var daoOnce = sync.Once{}

func InitCollectionOnce(db *mongox.Mongo) {
	daoOnce.Do(func() {
		err := dao.InitIndexes(db)
		if err != nil {
			panic(err)
		}
	})
}

func InitTaskDAO(db *mongox.Mongo) dao.TaskDAO {
	InitCollectionOnce(db)
	return dao.NewTaskDAO(db)
}

func initConsumer(svc service.Service, q mq.MQ) *event.ExecuteResultConsumer {
	consumer, err := event.NewExecuteResultConsumer(q, svc)
	if err != nil {
		panic(err)
	}

	consumer.Start(context.Background())
	return consumer
}

func InitCrypto(reg *cryptox.CryptoRegistry) cryptox.Crypto {
	return reg.Runner
}

func initStartTaskJob(svc service.Service) *StartTaskJob {
	limit := int64(100)
	initialInterval := 10 * time.Second
	maxInterval := 30 * time.Second
	maxRetries := int32(3)
	return job.NewStartTaskJob(svc, limit, initialInterval, maxInterval, maxRetries)
}

func initTaskRecoveryJob(svc service.Service) *TaskRecoveryJob {
	limit := int64(100)
	return job.NewTaskRecoveryJob(svc, limit)
}

func initPassProcessTaskJob(svc service.Service, engineSvc engine.Service) *PassProcessTaskJob {
	minutes := int64(10)
	seconds := int64(10)
	limit := int64(100)
	return job.NewPassProcessTaskJob(svc, engineSvc, minutes, seconds, limit)
}

func initTaskExecutionSyncJob(svc service.Service, engineSvc engine.Service, executorSvc executorv1.TaskExecutionServiceClient) *TaskExecutionSyncJob {
	limit := int64(100)
	return job.NewTaskExecutionSyncJob(svc, executorSvc, limit)
}
