//go:build wireinject

package task

import (
	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/runner"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository"
	"github.com/Duke1616/ecmdb/internal/task/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	service.NewService,
	repository.NewTaskRepository,
	dao.NewTaskDAO,
)

func InitModule(db *mongox.Mongo, orderModule *order.Module, workflowModule *workflow.Module,
	codebookModule *codebook.Module, workerModule *worker.Module, runnerModule *runner.Module) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.FieldsOf(new(*order.Module), "Svc"),
		wire.FieldsOf(new(*workflow.Module), "Svc"),
		wire.FieldsOf(new(*codebook.Module), "Svc"),
		wire.FieldsOf(new(*worker.Module), "Svc"),
		wire.FieldsOf(new(*runner.Module), "Svc"),

		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
