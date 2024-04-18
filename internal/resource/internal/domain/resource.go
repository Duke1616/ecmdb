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
