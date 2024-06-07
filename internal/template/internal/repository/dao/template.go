package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
)

type TemplateDAO interface {
	CreateTemplate(ctx context.Context, t Template) error
}

func NewTemplateDAO(db *mongox.Mongo) TemplateDAO {
	return &templateDAO{
		db: db,
	}
}

type templateDAO struct {
	db *mongox.Mongo
}

func (dao *templateDAO) CreateTemplate(ctx context.Context, t Template) error {
	panic("implement me")
}

type Template struct {
	Id   int64  `bson:"id"`
	Name string `bson:"name"`
}
