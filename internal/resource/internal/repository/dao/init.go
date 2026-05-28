package dao

import (
	"context"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitIndexes(db *mongox.DB) error {
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
		// 使用 percona mongo 创建全文检索，ngram 进行分词
		{
			Keys:    bson.D{{Key: "$**", Value: "text"}},
			Options: options.Index().SetDefaultLanguage("ngram"),
		},
	}

	return mongox.SyncIndexes(ctx, col, indexes)
}
