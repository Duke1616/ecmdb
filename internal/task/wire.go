//go:build wireinject

package task

import (
	"context"
	"fmt"
	"time"

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
	"github.com/Duke1616/ecmdb/internal/task/internal/web"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/mq-api"
	"github.com/google/wire"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/spf13/viper"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewExecService,
	service.NewService,
	service.NewCronjob,
	repository.NewTaskRepository,
	dao.NewTaskDAO,
)

func InitModule(q mq.MQ, db *mongox.Mongo, orderModule *order.Module, workflowModule *workflow.Module,
	engineModule *engine.Module, codebookModule *codebook.Module, workerModule *worker.Module,
	runnerModule *runner.Module, userModule *user.Module, discoveryModule *discovery.Module,
	lark *lark.Client, crypto *cryptox.CryptoRegistry, sender sender.NotificationSender) (*Module, error) {
	wire.Build(
		ProviderSet,
		initStartTaskJob,
		initPassProcessTaskJob,
		initRecoveryTaskJob,
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

func initConsumer(svc service.Service, q mq.MQ, codebookSvc codebook.Service,
	userSvc user.Service, sender sender.NotificationSender) *event.ExecuteResultConsumer {
	consumer, err := event.NewExecuteResultConsumer(q, svc, codebookSvc, userSvc, sender)
	if err != nil {
		panic(err)
	}

	consumer.Start(context.Background())
	return consumer
}

func InitCrypto(reg *cryptox.CryptoRegistry) cryptox.Crypto[string] {
	return reg.Runner
}

func initStartTaskJob(svc service.Service) *StartTaskJob {
	limit := int64(100)
	initialInterval := 10 * time.Second
	maxInterval := 30 * time.Second
	maxRetries := int32(3)
	return job.NewStartTaskJob(svc, limit, initialInterval, maxInterval, maxRetries)
}

func initRecoveryTaskJob(svc service.Service, execSvc service.ExecService, jobSvc service.Cronjob) *RecoveryTaskJob {
	limit := int64(100)
	recovery := job.NewRecoveryTaskJob(svc, execSvc, jobSvc, limit)

	type Config struct {
		Enabled bool `mapstructure:"enabled"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("cronjob", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into structure: %v", err))
	}

	if !cfg.Enabled {
		return recovery
	}

	go func() {
		_ = recovery.Run(context.Background())
	}()

	return nil
}

func initPassProcessTaskJob(svc service.Service, engineSvc engine.Service) *PassProcessTaskJob {
	minutes := int64(10)
	seconds := int64(10)
	limit := int64(100)
	return job.NewPassProcessTaskJob(svc, engineSvc, minutes, seconds, limit)
}
