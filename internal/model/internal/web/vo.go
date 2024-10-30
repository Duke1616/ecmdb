package web

import (
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
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

type RetrieveModelGroupsListResp struct {
	Total int64        `json:"total,omitempty"`
	Mgs   []ModelGroup `json:"model_groups,omitempty"`
}

type CreateModelRelationReq struct {
	SourceModelUID  string `json:"source_model_uid"`
	TargetModelUID  string `json:"target_model_uid"`
	RelationTypeUID string `json:"relation_type_uid"`
	Mapping         string `json:"mapping"`
}

type ModelGroup struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}
type Model struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	UID  string `json:"uid"`
	Icon string `json:"icon"`
	// 记录模型下面有多少资产
	Total int    `json:"total"`
	Ctime string `json:"ctime"`
	Utime string `json:"utime"`
}

type ModelRelation struct {
	ID              int64  `json:"id"`
	RelationTypeUID string `json:"relation_type_uid"`
	TargetModelUID  string `json:"target_model_uid"`
}

type RetrieveRelationModelGraph struct {
	RootId string      `json:"rootId"`
	Nodes  []ModelNode `json:"nodes"`
	Lines  []ModelLine `json:"lines"`
}

type ModelNode struct {
	ID   string            `json:"id"`
	Text string            `json:"text"`
	Data map[string]string `json:"data,omitempty"`
}

type DeleteModelByUidReq struct {
	ModelUid string `json:"model_uid"`
}

type DeleteModelGroup struct {
	ID int64 `json:"id"`
}

type ModelLine struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}

func toModelVo(m domain.Model) Model {
	return Model{
		Name:  m.Name,
		UID:   m.UID,
		Ctime: m.Utime.Format(time.DateTime),
		Utime: m.Utime.Format(time.DateTime),
	}
}
