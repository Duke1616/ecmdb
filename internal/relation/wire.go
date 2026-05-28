//go:build wireinject

package relation

import (
	"sync"

	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/internal/relation/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	web.NewRelationResourceHandler,
	web.NewRelationModelHandler,
	web.NewRelationTypeHandler,
	service.NewRelationTypeService,
	repository.NewRelationTypeRepository,
	repository.NewRelationModelRepository,
	repository.NewRelationResourceRepository,
	initRmDAO,
	initRrDAO,
)

func InitModule(db *mongox.DB) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitRelationTypeDAO,
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

func initRmDAO(db *mongox.DB) dao.RelationModelDAO {
	rmDaoOnce.Do(func() {
		rmd = dao.NewRelationModelDAO(db)
	})
	return rmd
}

var daoOnce = sync.Once{}

func InitCollectionOnce(db *mongox.DB) {
	daoOnce.Do(func() {
		err := dao.InitIndexes(db)
		if err != nil {
			panic(err)
		}
	})
}

func InitRelationTypeDAO(db *mongox.DB) dao.RelationTypeDAO {
	InitCollectionOnce(db)
	return dao.NewRelationTypeDAO(db)
}

func InitRMService(db *mongox.DB) RMSvc {
	wire.Build(
		initRmDAO,
		initRrDAO,
		repository.NewRelationModelRepository,
		repository.NewRelationResourceRepository,
		service.NewRelationModelService,
	)
	return nil
}

func initRrDAO(db *mongox.DB) dao.RelationResourceDAO {
	rrDaoOnce.Do(func() {
		rrd = dao.NewRelationResourceDAO(db)
	})
	return rrd
}

func InitRRService(db *mongox.DB) RRSvc {
	wire.Build(
		initRrDAO,
		initRmDAO,
		repository.NewRelationResourceRepository,
		repository.NewRelationModelRepository,
		service.NewRelationResourceService,
	)
	return nil
}
