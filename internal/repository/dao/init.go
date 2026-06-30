package dao

import (
	"context"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitIndexes(db *mongox.DB) error {
	// Model 索引
	if err := InitModelIndexes(db); err != nil {
		return err
	}
	if err := initModelGroupIndex(db); err != nil {
		return err
	}

	// Attribute 索引
	if err := initAttrIndex(db); err != nil {
		return err
	}
	if err := initAttrGroupIndex(db); err != nil {
		return err
	}

	// Resource 索引
	if err := initResourceIndexes(db); err != nil {
		return err
	}

	// Relation 索引
	if err := initRTIndex(db); err != nil {
		return err
	}
	if err := initRMIndex(db); err != nil {
		return err
	}

	// Plugin 索引
	if err := initPluginIndexes(db); err != nil {
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

func initPluginIndexes(db *mongox.DB) error {
	ctx := context.Background()

	if err := mongox.SyncIndexes(ctx, db.Database().Collection(PluginCollection), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "uid", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}); err != nil {
		return err
	}

	return mongox.SyncIndexes(ctx, db.Database().Collection(PluginBindingCollection), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "uid", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "model_uid", Value: 1},
				{Key: "enabled", Value: 1},
			},
		},
	})
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

func initAttrIndex(db *mongox.DB) error {
	col := mongox.NewCollection[Attribute](db, AttributeCollection)
	ctx := context.Background()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
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

	return mongox.SyncIndexes(ctx, col.Native(), indexes)
}

func initAttrGroupIndex(db *mongox.DB) error {
	col := mongox.NewCollection[AttributeGroup](db, AttributeGroupCollection)
	ctx := context.Background()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "model_uid", Value: -1},
				{Key: "name", Value: -1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	return mongox.SyncIndexes(ctx, col.Native(), indexes)
}

func initResourceIndexes(db *mongox.DB) error {
	col := db.Database().Collection(ResourceCollection)
	ctx := context.Background()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "name", Value: -1},
				{Key: "model_uid", Value: -1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "$**", Value: "text"}},
			Options: options.Index().SetDefaultLanguage("ngram"),
		},
	}

	return mongox.SyncIndexes(ctx, col, indexes)
}

func initRTIndex(db *mongox.DB) error {
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
