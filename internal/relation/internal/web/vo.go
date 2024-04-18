package web

import (
	"time"
)

// CreateModelRelationReq 模型关联关系
type CreateModelRelationReq struct {
	SourceModelUID  string `json:"source_model_uid"`
	TargetModelUID  string `json:"target_model_uid"`
	RelationTypeUID string `json:"relation_type_uid"`
	Mapping         string `json:"mapping"`
}

// CreateResourceRelationReq 资源关联关系
type CreateResourceRelationReq struct {
	SourceModelUID   string `json:"source_model_uid"`
	TargetModelUID   string `json:"target_model_uid"`
	SourceResourceID int64  `json:"source_resource_id"`
	TargetResourceID int64  `json:"target_resource_id"`
	RelationTypeUID  string `json:"relation_type_uid"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListModelRelationByModelUidReq struct {
	Page
	ModelUid string `json:"model_uid"`
}

type ListResourceRelationByModelUidReq struct {
	Page
	ID           string `json:"id"`
	ModelUid     string `json:"model_uid"`
	RelationType string `json:"relation_type"`
}

type ListRelationModelsResp struct {
	Total          int64           `json:"total,omitempty"`
	ModelRelations []ModelRelation `json:"orders,omitempty"`
}

type ModelRelation struct {
	ID              int64     `json:"id"`
	SourceModelUID  string    `json:"source_model_uid"`
	TargetModelUID  string    `json:"target_model_uid"`
	RelationTypeUID string    `json:"relation_type_uid"` // 关联类型唯一索引
	RelationName    string    `json:"relation_name"`     // 拼接字符
	Mapping         string    `json:"mapping"`           // 关联关系
	Ctime           time.Time `json:"ctime"`
	Utime           time.Time `json:"utime"`
}

type CreateRelationTypeReq struct {
	Name           string `json:"name"`
	UID            string `json:"uid"`
	SourceDescribe string `json:"source_describe"`
	TargetDescribe string `json:"target_describe"`
}

// Model 拓补图模型关联节点信息
type Model struct {
	ID              int64  `json:"id"`
	RelationTypeUID string `json:"relation_type_uid"`
	TargetModelUID  string `json:"target_model_uid"`
}

type ModelDiagram struct {
	ID        int64  `json:"id"`
	Icon      string `json:"icon"`
	ModelUID  string `json:"model_uid"`
	ModelName string `json:"model_name"`
	Assets    []Model
}

type RetrieveRelationModelDiagram struct {
	Diagrams []ModelDiagram `json:"diagrams"`
}

type ResourceRelation struct {
	SourceModelUID   string `json:"source_model_uid"`
	TargetModelUID   string `json:"target_model_uid"`
	SourceResourceID int64  `json:"source_resource_id"`
	TargetResourceID int64  `json:"target_resource_id"`
	RelationTypeUID  string `json:"relation_type_uid"`
	RelationName     string `json:"relation_name"`
}

type ListResourceDiagramReq struct {
	ModelUid   string `json:"model_uid"`
	ResourceId int64  `json:"resource_id"`
}

type RetrieveResource struct {
	Name   string             `json:"name"`
	Assets []ResourceRelation `json:"assets"`
}

type ListModelByUidReq struct {
	ModelUid string `json:"model_uid"`
}

type Data struct {
	ModelUid         string `json:"model_uid"`
	ResourceId       int64  `json:"resource_id"`
	RelationTypeName string `json:"relation_type_name"`
}

type RetrieveDiagram struct {
	SRC    []ResourceRelation  `json:"src"`
	DST    []ResourceRelation  `json:"dst"`
	Assets map[string][]Assets `json:"assets"`
}

type Assets struct {
	ResourceID   int64  `json:"resource_id"`
	ResourceName string `json:"resource_name"`
}
