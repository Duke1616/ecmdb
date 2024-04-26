package web

import "github.com/Duke1616/ecmdb/pkg/mongox"

type CreateResourceReq struct {
	Name     string        `json:"name"`
	ModelUid string        `json:"model_uid"`
	Data     mongox.MapStr `json:"data"`
}

type DetailResourceReq struct {
	ModelUid string `json:"model_uid"`
	ID       int64  `json:"id"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListResourceReq struct {
	Page
	ModelUid string `json:"model_uid"`
}

type ListResourceIdsReq struct {
	ModelUid string  `json:"model_uid"`
	Ids      []int64 `json:"ids"`
}

// ListCanBeRelatedReq 查询可以关联的节点
type ListCanBeRelatedReq struct {
	Page
	ResourceId   int64  `json:"resource_id"`   // 当前资源ID
	ModelUid     string `json:"model_uid"`     // 当前模型ID
	RelationName string `json:"relation_name"` // 关联类型，以方便推断是数据正向 OR 反向
}

type ListDiagramReq struct {
	ModelUid   string `json:"model_uid"`
	ResourceId int64  `json:"resource_id"`
}

type ResourceRelation struct {
	SourceModelUID   string `json:"source_model_uid"`
	TargetModelUID   string `json:"target_model_uid"`
	SourceResourceID int64  `json:"source_resource_id"`
	TargetResourceID int64  `json:"target_resource_id"`
	RelationTypeUID  string `json:"relation_type_uid"`
	RelationName     string `json:"relation_name"`
}

type ResourceAssets struct {
	ResourceID   int64  `json:"resource_id"`
	ResourceName string `json:"resource_name"`
}

type RetrieveDiagram struct {
	SRC    []ResourceRelation          `json:"src"`
	DST    []ResourceRelation          `json:"dst"`
	Assets map[string][]ResourceAssets `json:"assets"`
}
