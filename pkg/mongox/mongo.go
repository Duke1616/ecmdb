package mongox

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	DBClient *mongo.Client
	Sess     mongo.Session
	dbName   string
}

func NewMongo(client *mongo.Client, dbName string) *Mongo {
	return &Mongo{
		DBClient: client,
		dbName:   dbName,
	}
}

func (m *Mongo) Database() *mongo.Database {
	return m.DBClient.Database(m.dbName)
}

func (m *Mongo) Collection(collName string) *mongo.Collection {
	return m.Database().Collection(collName)
}

func (m *Mongo) Collections(collName string) Collection {
	col := Coll{}
	col.collName = collName
	return &col
}

func (m *Mongo) GetIdGenerator(collection string) int64 {
	coll := m.Database().Collection("c_id_generator")
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
