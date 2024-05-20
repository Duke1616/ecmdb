package web

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

// CreateRelationTypeReq 关联关系类型
type CreateRelationTypeReq struct {
	Name           string `json:"name"`
	UID            string `json:"uid"`
	SourceDescribe string `json:"source_describe"`
	TargetDescribe string `json:"target_describe"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListModelRelationReq struct {
	Page
	ModelUid string `json:"model_uid"`
}

type ListResourceDiagramReq struct {
	ModelUid   string `json:"model_uid"`
	ResourceId int64  `json:"resource_id"`
}

type RetrieveRelationModels struct {
	Total          int64           `json:"total,omitempty"`
	ModelRelations []ModelRelation `json:"model_relations,omitempty"`
}

type RetrieveRelationType struct {
	Total         int64          `json:"total,omitempty"`
	RelationTypes []RelationType `json:"relation_types,omitempty"`
}

type RetrieveAggregatedAssets struct {
	RelationName string  `json:"relation_name"`
	ModelUid     string  `json:"model_uid"`
	Total        int     `json:"total"`
	ResourceIds  []int64 `json:"resource_ids"`
}

type RetrieveRelationResource struct {
	Total             int64              `json:"total,omitempty"`
	ResourceRelations []ResourceRelation `json:"resource_relations,omitempty"`
}

type DeleteModelRelationReq struct {
	Id int64 `json:"id"`
}

type DeleteResourceRelationReq struct {
	ModelUid     string `json:"model_uid"`
	ResourceId   int64  `json:"resource_id"`
	RelationName string `json:"relation_name"`
}
