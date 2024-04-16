package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/mongo"
)

type RelationTypeDAO interface {
	CreateRelation(ctx context.Context, mg ModelRelation) (int64, error)
}

func NewRelationTypeDAO(client *mongo.Client) RelationTypeDAO {
	return &relationTypeDAO{
		db: mongox.NewMongo(client),
	}
}

type relationTypeDAO struct {
	db *mongox.Mongo
}

func (r relationTypeDAO) CreateRelation(ctx context.Context, mg ModelRelation) (int64, error) {
	//TODO implement me
	panic("implement me")
}
