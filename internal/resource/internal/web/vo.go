package web

import "github.com/Duke1616/ecmdb/pkg/mongox"

type CreateResourceReq struct {
	Data mongox.MapStr `json:"data"`
}

type DetailResourceReq struct {
	ID         int64       `json:"id"`
	Attributes []Attribute `json:"attributes,omitempty"`
}

type Attribute struct {
	ID         int64
	ModelID    int64
	Identifies string
	Name       string
	FieldType  string
	Required   bool
}
