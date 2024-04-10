package domain

import "github.com/Duke1616/ecmdb/pkg/mongox"

type Resource struct {
	ID      int64
	ModelID int64
	Data    mongox.MapStr
}
