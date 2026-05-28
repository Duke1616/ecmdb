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
	ctx := context.Background()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "uid", Value: -1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	return mongox.SyncIndexes(ctx, col, indexes)
}

func initModelGroupIndex(db *mongox.DB) error {
	col := db.Database().Collection(ModelGroupCollection)
	ctx := context.Background()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "name", Value: -1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	return mongox.SyncIndexes(ctx, col, indexes)
}
