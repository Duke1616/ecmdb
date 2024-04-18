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

type ListRelationsReq struct {
	ModelUid string `json:"model_uid"`
}
