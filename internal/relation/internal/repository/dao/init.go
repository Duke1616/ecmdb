package dao

import (
	"context"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitIndexes(db *mongox.DB) error {
	if err := initRTIndex(db); err != nil {
		return err
	}

	if err := initRMIndex(db); err != nil {
		return err
	}

	return nil
}

func initRTIndex(db *mongox.DB) error {
	// 使用 Collection[RelationType].Native() 拿到底层原始驱动连接以安全操作 Index
	col := mongox.NewCollection[RelationType](db, RelationTypeCollection)
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

	return mongox.SyncIndexes(ctx, col.Native(), indexes)
}

func initRMIndex(db *mongox.DB) error {
	// 使用 Collection[ModelRelation].Native() 拿到底层原始驱动连接以安全操作 Index
	col := mongox.NewCollection[ModelRelation](db, ModelRelationCollection)
	ctx := context.Background()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "relation_name", Value: -1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	return mongox.SyncIndexes(ctx, col.Native(), indexes)
}
