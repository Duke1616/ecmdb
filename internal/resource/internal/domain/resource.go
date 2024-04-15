package domain

import "github.com/Duke1616/ecmdb/pkg/mongox"

type Resource struct {
	ID       int64
	Name     string
	ModelUID string // 因为这个传参是 URL PATH, 使用ID会显得丑陋，所以使用模型唯一身份标识
	Data     mongox.MapStr
}

type ResourceRelation struct {
	ModelUid  string
	Resources []Resource
}

type DetailResource struct {
	ID         int64 `json:"id"`
	Projection map[string]int
}
