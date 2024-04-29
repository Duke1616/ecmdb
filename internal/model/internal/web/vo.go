package web

import (
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type CreateModelGroupReq struct {
	Name string `json:"name"`
}

type CreateModelReq struct {
	Name    string `json:"name"`
	GroupId int64  `json:"group_id"`
	UID     string `json:"uid"`
	Icon    string `json:"icon"`
}

type DetailModelReq struct {
	ID int64 `json:"id"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type RetrieveModelsListResp struct {
	Total  int64   `json:"total,omitempty"`
	Models []Model `json:"models,omitempty"`
}

type CreateModelRelationReq struct {
	SourceModelUID  string `json:"source_model_uid"`
	TargetModelUID  string `json:"target_model_uid"`
	RelationTypeUID string `json:"relation_type_uid"`
	Mapping         string `json:"mapping"`
}

type Model struct {
	Name  string `json:"name"`
	UID   string `json:"uid"`
	Icon  string `json:"icon"`
	Ctime string `json:"ctime"`
	Utime string `json:"utime"`
}

type ModelRelation struct {
	ID              int64  `json:"id"`
	RelationTypeUID string `json:"relation_type_uid"`
	TargetModelUID  string `json:"target_model_uid"`
}

// RelationModel 拓补图模型关联节点信息
type RelationModel struct {
	ID              int64  `json:"id"`
	RelationTypeUID string `json:"relation_type_uid"`
	TargetModelUID  string `json:"target_model_uid"`
}

type ModelDiagram struct {
	ID        int64           `json:"id"`
	Icon      string          `json:"icon"`
	ModelUID  string          `json:"model_uid"`
	ModelName string          `json:"model_name"`
	Assets    []RelationModel `json:"assets"`
}

type RetrieveRelationModelDiagram struct {
	Diagrams []ModelDiagram `json:"diagrams"`
}

func toModelVo(m domain.Model) Model {
	return Model{
		Name:  m.Name,
		UID:   m.UID,
		Ctime: m.Utime.Format(time.DateTime),
		Utime: m.Utime.Format(time.DateTime),
	}
}

func toModelDiagramVo(models []domain.Model, mds map[string][]relation.ModelDiagram) []ModelDiagram {
	return slice.Map(models, func(idx int, src domain.Model) ModelDiagram {
		var m []RelationModel
		val, ok := mds[src.UID]
		if ok {
			m = slice.Map(val, func(idx int, src relation.ModelDiagram) RelationModel {
				return RelationModel{
					ID:              src.ID,
					RelationTypeUID: src.RelationTypeUid,
					TargetModelUID:  src.TargetModelUid,
				}
			})
		}

		return ModelDiagram{
			ID:        src.ID,
			Icon:      src.Icon,
			ModelUID:  src.UID,
			ModelName: src.Name,
			Assets:    m,
		}
	})
}
