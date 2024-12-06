package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitIndexes(db *mongox.Mongo) error {
	col := db.Collection(ResourceCollection)

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{"name", -1},
				{"model_uid", -1},
			},
			Options: options.Index().SetUnique(true),
		},
		// 使用 percona mongo 创建全文检索，ngram 进行分词
		{
			Keys:    bson.D{{Key: "$**", Value: "text"}},
			Options: options.Index().SetDefaultLanguage("ngram"),
		},
	}

	_, err := col.Indexes().CreateMany(context.Background(), indexes)

	return err
}
