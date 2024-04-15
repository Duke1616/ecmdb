package web

import "github.com/Duke1616/ecmdb/pkg/mongox"

type CreateResourceReq struct {
	Name string        `json:"name"`
	Data mongox.MapStr `json:"data"`
}

type DetailResourceReq struct {
	ID int64 `json:"id"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListRelationReq struct {
}
