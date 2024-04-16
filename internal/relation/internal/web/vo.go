package web

import "time"

type CreateModelRelationReq struct {
	SourceModelUID  string `json:"source_model_uid"`
	TargetModelUID  string `json:"target_model_uid"`
	RelationTypeUID string `json:"relation_type_uid"`
	Mapping         string `json:"mapping"`
}

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
	ModelUid     string `json:"model_uid"`
	RelationType string `json:"relation_type"`
}

type ListOrdersResp struct {
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
