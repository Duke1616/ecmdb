package dao

import (
	"context"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitIndexes(db *mongox.Mongo) error {
	if err := initRTIndex(db); err != nil {
		return err
	}

	if err := initRMIndex(db); err != nil {
		return err
	}

	return nil
}

func initRTIndex(db *mongox.Mongo) error {
	col := db.Collection(RelationTypeCollection)

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.M{"uid": -1},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := col.Indexes().CreateMany(context.Background(), indexes)

	return err
}

func initRMIndex(db *mongox.Mongo) error {
	col := db.Collection(ModelRelationCollection)

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.M{"relation_name": -1},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := col.Indexes().CreateMany(context.Background(), indexes)

	return err
}
