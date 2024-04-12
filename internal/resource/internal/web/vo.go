package web

import "github.com/Duke1616/ecmdb/pkg/mongox"

type CreateResourceReq struct {
	Data mongox.MapStr `json:"data"`
}

type DetailResourceReq struct {
	ID int64 `json:"id"`
}
