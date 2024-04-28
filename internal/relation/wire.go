//go:build wireinject

package relation

import (
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/internal/relation/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
	"sync"
)

var ProviderSet = wire.NewSet(
	web.NewRelationResourceHandler,
	web.NewRelationModelHandler,
	web.NewRelationTypeHandler,
	service.NewRelationTypeService,
	repository.NewRelationTypeRepository,
	dao.NewRelationTypeDAO)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitRRService,
		InitRMService,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

var (
	rmDaoOnce = sync.Once{}
	rrDaoOnce = sync.Once{}
	rmd       dao.RelationModelDAO
	rrd       dao.RelationResourceDAO
)

func initRmDAO(db *mongox.Mongo) dao.RelationModelDAO {
	rmDaoOnce.Do(func() {
		rmd = dao.NewRelationModelDAO(db)
	})
	return rmd
}

func InitRMService(db *mongox.Mongo) RMSvc {
	wire.Build(
		initRmDAO,
		repository.NewRelationModelRepository,
		service.NewRelationModelService,
	)
	return nil
}

func intRrDAO(db *mongox.Mongo) dao.RelationResourceDAO {
	rrDaoOnce.Do(func() {
		rrd = dao.NewRelationResourceDAO(db)
	})
	return rrd
}

func InitRRService(db *mongox.Mongo) RRSvc {
	wire.Build(
		intRrDAO,
		repository.NewRelationResourceRepository,
		service.NewRelationResourceService,
	)
	return nil
}
