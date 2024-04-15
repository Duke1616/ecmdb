package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitIndexes(db *mongo.Client) error {
	col := mongox.NewMongo(db).Collection(ResourceCollection)

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{"name", -1},
				{"model_uid", -1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := col.Indexes().CreateMany(context.Background(), indexes)

	return err
}
