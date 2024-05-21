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

type ListResourceByIdsReq struct {
	ModelUid    string  `json:"model_uid"`
	ResourceIds []int64 `json:"resource_ids"`
}

type DeleteResourceReq struct {
	Id int64 `json:"id"`
}

// ListCanBeRelatedReq 查询可以关联的节点
type ListCanBeRelatedReq struct {
	Page
	ResourceId   int64  `json:"resource_id"`   // 当前资源ID
	ModelUid     string `json:"model_uid"`     // 当前模型ID
	RelationName string `json:"relation_name"` // 关联类型，以方便推断是数据正向 OR 反向
}

type ListDiagramReq struct {
	ModelUid     string `json:"model_uid"`
	ResourceId   int64  `json:"resource_id"`
	ResourceName string `json:"resource_name"`
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

type RetrieveGraph struct {
	RootId string `json:"rootId"`
	Nodes  []Node `json:"nodes"`
	Lines  []Line `json:"lines"`
}

type Node struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	// 扩展方向
	ExpandHolderPosition string            `json:"expandHolderPosition,omitempty"`
	Data                 map[string]string `json:"data,omitempty"`
}

type Line struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Resource struct {
	ID       int64         `json:"id"`
	Name     string        `json:"name"`
	ModelUID string        `json:"model_uid"`
	Data     mongox.MapStr `json:"data"`
}

type RetrieveResources struct {
	Resources []Resource `json:"resources"`
	Total     int64      `json:"total"`
}
