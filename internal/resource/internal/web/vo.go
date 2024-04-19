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

// ListRelatedReq 查询指定关联的数据
// 根据传入模型以及关联名称，推断出对方的模型，排除已经关联数据，返回对应的数据
type ListRelatedReq struct {
	Page
	ResourceId   int64  `json:"resource_id"`   // 当前资源ID
	ModelUid     string `json:"model_uid"`     // 当前模型ID
	RelationName string `json:"relation_name"` // 关联类型，以方便推断是数据正向 OR 反向
}
