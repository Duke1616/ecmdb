package dao

import (
	"context"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitIndexes(db *mongox.Mongo) error {
	if err := initAttrIndex(db); err != nil {
		return err
	}

	if err := initAttrGroupIndex(db); err != nil {
		return err
	}

	return nil
}

func initAttrIndex(db *mongox.Mongo) error {
	col := db.Collection(AttributeCollection)

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "field_uid", Value: -1},
				{Key: "model_uid", Value: -1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "model_uid", Value: 1},
				{Key: "sort_key", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "group_id", Value: 1},
				{Key: "sort_key", Value: 1},
			},
		},
	}

	_, err := col.Indexes().CreateMany(context.Background(), indexes)

	return err
}

func initAttrGroupIndex(db *mongox.Mongo) error {
	col := db.Collection(AttributeGroupCollection)

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "model_uid", Value: -1},
				{Key: "name", Value: -1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := col.Indexes().CreateMany(context.Background(), indexes)

	return err
}
