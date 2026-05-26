package dao

import (
	"context"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitIndexes(db *mongox.DB) error {
	if err := InitModelIndexes(db); err != nil {
		return err
	}

	if err := initModelGroupIndex(db); err != nil {
		return err
	}

	return nil
}

func InitModelIndexes(db *mongox.DB) error {
	col := db.Database().Collection(ModelCollection)

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.M{"uid": -1},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := col.Indexes().CreateMany(context.Background(), indexes)

	return err
}

func initModelGroupIndex(db *mongox.DB) error {
	col := db.Database().Collection(ModelGroupCollection)

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{"name", -1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := col.Indexes().CreateMany(context.Background(), indexes)

	return err
}
