package relation

import (
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/internal/relation/internal/web"
)

// RR => RelationResource
// RM => RelationModel
// RT => RelationType

type RRSvc = service.RelationResourceService
type RRHandler = web.RelationResourceHandler

type RMSvc = service.RelationModelService
type RMHandler = web.RelationModelHandler

type RTSvc = service.RelationTypeService
type RTHandler = web.RelationTypeHandler

type ModelDiagram = domain.ModelDiagram
