// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package task

import (
	"context"
	"fmt"
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
	"github.com/larksuite/oapi-sdk-go/v3"
	"github.com/spf13/viper"
	"time"
)

// Injectors from wire.go:

func InitModule(q mq.MQ, db *mongox.Mongo, orderModule *order.Module, workflowModule *workflow.Module, engineModule *engine.Module, codebookModule *codebook.Module, workerModule *worker.Module, runnerModule *runner.Module, userModule *user.Module, lark2 *lark.Client) (*Module, error) {
	taskDAO := dao.NewTaskDAO(db)
	taskRepository := repository.NewTaskRepository(taskDAO)
	serviceService := orderModule.Svc
	service2 := workflowModule.Svc
	service3 := codebookModule.Svc
	service4 := runnerModule.Svc
	service5 := workerModule.Svc
	execService := service.NewExecService(service5)
	cronjob := service.NewCronjob(execService)
	service6 := engineModule.Svc
	service7 := userModule.Svc
	service8 := service.NewService(taskRepository, serviceService, service2, service3, service4, cronjob, service6, service7, execService)
	handler := web.NewHandler(service8)
	executeResultConsumer := initConsumer(service8, q, service3, service7, lark2)
	startTaskJob := initStartTaskJob(service8)
	passProcessTaskJob := initPassProcessTaskJob(service8, service6)
	recoveryTaskJob := initRecoveryTaskJob(service8, execService, cronjob)
	module := &Module{
		Svc:                service8,
		Hdl:                handler,
		c:                  executeResultConsumer,
		StartTaskJob:       startTaskJob,
		PassProcessTaskJob: passProcessTaskJob,
		RecoveryTaskJob:    recoveryTaskJob,
	}
	return module, nil
}

// wire.go:

var ProviderSet = wire.NewSet(web.NewHandler, service.NewExecService, service.NewService, service.NewCronjob, repository.NewTaskRepository, dao.NewTaskDAO)

func initConsumer(svc service.Service, q mq.MQ, codebookSvc codebook.Service,
	userSvc user.Service, lark2 *lark.Client) *event.ExecuteResultConsumer {
	consumer, err := event.NewExecuteResultConsumer(q, svc, codebookSvc, userSvc, lark2)
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

func initRecoveryTaskJob(svc service.Service, execSvc service.ExecService, jobSvc service.Cronjob) *RecoveryTaskJob {
	limit := int64(100)
	recovery := job.NewRecoveryTaskJob(svc, execSvc, jobSvc, limit)

	type Config struct {
		Enabled bool `mapstructure:"enabled"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("cronjob", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %v", err))
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
	minutes := int64(60)
	seconds := int64(10)
	limit := int64(100)
	return job.NewPassProcessTaskJob(svc, engineSvc, minutes, seconds, limit)
}
