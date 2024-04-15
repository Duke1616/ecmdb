package web

import "time"

type CreateRelationTypeReq struct {
	name string
}

type CreateModelRelationReq struct {
	SourceModelIdentifies  string `json:"source_model_identifies"`
	TargetModelIdentifies  string `json:"target_model_identifies"`
	RelationTypeIdentifies string `json:"relation_type_identifies"`
	Mapping                string `json:"mapping"`
}

type CreateResourceRelationReq struct {
	SourceModelIdentifies  string `json:"source_model_identifies"`
	TargetModelIdentifies  string `json:"target_model_identifies"`
	SourceResourceID       int64  `json:"source_resource_id"`
	TargetResourceID       int64  `json:"target_resource_id"`
	RelationTypeIdentifies string `json:"relation_type_identifies"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListModelRelationByModelIdentifiesReq struct {
	Page
	ModelIdentifies string `json:"model_identifies"`
}

type ListOrdersResp struct {
	Total          int64           `json:"total,omitempty"`
	ModelRelations []ModelRelation `json:"orders,omitempty"`
}

type ModelRelation struct {
	ID                     int64     `json:"id"`
	SourceModelIdentifies  string    `json:"source_model_identifies"`
	TargetModelIdentifies  string    `json:"target_model_identifies"`
	RelationTypeIdentifies string    `json:"relation_type_identifies"` // 关联类型唯一索引
	RelationName           string    `json:"relation_name"`            // 拼接字符
	Mapping                string    `json:"mapping"`                  // 关联关系
	Ctime                  time.Time `json:"ctime"`
	Utime                  time.Time `json:"utime"`
}
