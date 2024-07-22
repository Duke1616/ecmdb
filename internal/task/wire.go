//go:build wireinject

package task

import (
	"github.com/Duke1616/ecmdb/internal/codebook"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/runner"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/Duke1616/ecmdb/internal/worker"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/google/wire"
)

func InitModule(orderModule *order.Module, workflowModule *workflow.Module, codebookModule *codebook.Module,
	workerModule *worker.Module, runnerModule *runner.Module) (*Module, error) {
	wire.Build(
		service.NewService,
		wire.FieldsOf(new(*order.Module), "Svc"),
		wire.FieldsOf(new(*workflow.Module), "Svc"),
		wire.FieldsOf(new(*codebook.Module), "Svc"),
		wire.FieldsOf(new(*worker.Module), "Svc"),
		wire.FieldsOf(new(*runner.Module), "Svc"),

		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
