package dao

import (
	"context"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitIndexes(db *mongox.Mongo) error {
	col := db.Collection(AttributeCollection)

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{"field_uid", -1},
				{"model_uid", -1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"model_uid": -1},
		},
	}

	_, err := col.Indexes().CreateMany(context.Background(), indexes)

	return err
}
