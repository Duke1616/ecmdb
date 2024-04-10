package web

import "github.com/Duke1616/ecmdb/pkg/mongox"

type CreateResourceReq struct {
	Data mongox.MapStr `json:"data"`
}
