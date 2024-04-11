package domain

import "github.com/Duke1616/ecmdb/pkg/mongox"

type Resource struct {
	ID              int64
	ModelIdentifies string
	Data            mongox.MapStr
}
