//go:build wireinject

package codebook

import (
	repository "github.com/Duke1616/ecmdb/internal/codebook/internal/repostory"
	"github.com/Duke1616/ecmdb/internal/codebook/internal/repostory/dao"
	"github.com/Duke1616/ecmdb/internal/codebook/internal/service"
	"github.com/Duke1616/ecmdb/internal/codebook/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/mq-api"
	"github.com/xen0n/go-workwx"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewCodebookRepository,
	dao.NewCodebookDAO)

func InitModule(q mq.MQ, db *mongox.Mongo, workAPP *workwx.WorkwxApp) (*Module, error) {
	wire.Build(
		ProviderSet,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}
