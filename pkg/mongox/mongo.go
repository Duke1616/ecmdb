package mongox

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetDataID(DB *mongo.Database, collection string) int64 {
	coll := DB.Collection("c_id_generator")
	var result struct {
		Name   string `json:"name" bson:"name"`
		NextID int64  `json:"next_id" bson:"next_id"`
	}

	update := bson.M{
		"$inc": bson.M{"next_id": int64(1)},
	}
	filter := bson.M{"name": collection}

	upsert := true
	returnChange := options.After
	opt := &options.FindOneAndUpdateOptions{
		Upsert:         &upsert,
		ReturnDocument: &returnChange,
	}

	err := coll.FindOneAndUpdate(context.Background(), filter, update, opt).Decode(&result)
	if err != nil {
		return 0
	}

	return result.NextID
}
