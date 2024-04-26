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
	SourceResourceID int64  `json:"source_resource_id"`
	TargetResourceID int64  `json:"target_resource_id"`
	RelationName     string `json:"relation_name"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListModelRelationReq struct {
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

// ListRelatedReq 查询指定关联的数据
// 根据传入模型以及关联名称，推断出对方的模型，排除已经关联数据，返回对应的数据
type ListRelatedReq struct {
	Page
	ResourceId   int64  `json:"resource_id"`   // 当前资源ID
	ModelUid     string `json:"model_uid"`     // 当前模型ID
	RelationName string `json:"relation_name"` // 关联类型，以方便推断是数据正向 OR 反向
}
