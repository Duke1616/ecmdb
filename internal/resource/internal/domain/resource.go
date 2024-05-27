package domain

import "github.com/Duke1616/ecmdb/pkg/mongox"

type Resource struct {
	ID       int64         `json:"id"`
	Name     string        `json:"name"`
	ModelUID string        `json:"model_uid"`
	Data     mongox.MapStr `json:"data"`
}

type ResourceRelation struct {
	ModelUid  string
	Resources []Resource
}

type SearchResource struct {
	ModelUid string
	Total    int `json:"total"`
	Data     []mongox.MapStr
}

type Condition struct {
	Name      string `json:"name"`      // 过滤名称
	Condition string `json:"condition"` // 过滤条件
	Input     string `json:"input"`     // 过滤输入
}
