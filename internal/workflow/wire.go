//go:build wireinject

package workflow

import (
	"github.com/Duke1616/ecmdb/internal/workflow/internal/repository"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/service"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/web"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewWorkflowRepository,
	dao.NewWorkflowDAO,
	easyflow.NewLogicFlowToEngineConvert,
)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
