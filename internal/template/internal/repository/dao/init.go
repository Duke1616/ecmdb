package dao

import (
	"context"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitIndexes(db *mongox.Mongo) error {
	// Template 集合索引
	if err := initTemplateIndexes(db); err != nil {
		return err
	}

	// TemplateFavorite 集合索引
	if err := initFavoriteIndexes(db); err != nil {
		return err
	}

	return nil
}

func initTemplateIndexes(db *mongox.Mongo) error {
	col := db.Collection(TemplateCollection)
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "workflow_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "unique_hash", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "external_template_id", Value: 1}},
		},
	}
	_, err := col.Indexes().CreateMany(context.Background(), indexes)
	return err
}

func initFavoriteIndexes(db *mongox.Mongo) error {
	col := db.Collection(TemplateFavoriteCollection)
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		// user_id 和 template_id 联合唯一索引，用于处理幂等收藏操作
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "template_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		// 单独 user_id 索引，方便根据用户查找收藏列表
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
	}
	_, err := col.Indexes().CreateMany(context.Background(), indexes)
	return err
}
