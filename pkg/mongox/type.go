package mongox

import "context"

type MapStr map[string]interface{}

type Filter interface{}

type Querier[T any] interface {
	FindOne(ctx context.Context) (*T, error)
	FindMany(ctx context.Context) ([]*T, error)
}
