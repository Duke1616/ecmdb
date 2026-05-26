package mongox

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ============================================================================
// 旧版兼容层（向下兼容未重构模块，避免大面积编译故障）
// ============================================================================

// AutoIDModel 定义需要自动生成 ID 的模型接口 (兼容旧版)
type AutoIDModel interface {
	SetID(id int64)
	GetID() int64
}

// Mongo 旧版 Mongo 核心处理器
type Mongo struct {
	DBClient *mongo.Client
	Sess     mongo.Session
	dbName   string
}

// NewMongo 兼容旧版初始化 (注意：返回 *Mongo 而非泛型新版 *DB)
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

// GetBatchIdGenerator 批量获取自增 ID，提高批量插入性能
func (m *Mongo) GetBatchIdGenerator(collection string, count int) (startID int64, err error) {
	if count <= 0 {
		return 0, nil
	}

	coll := m.Database().Collection("c_id_generator")
	var result struct {
		Name   string `json:"name" bson:"name"`
		NextID int64  `json:"next_id" bson:"next_id"`
	}

	update := bson.M{
		"$inc": bson.M{"next_id": int64(count)},
	}
	filter := bson.M{"name": collection}

	upsert := true
	returnChange := options.After
	opt := &options.FindOneAndUpdateOptions{
		Upsert:         &upsert,
		ReturnDocument: &returnChange,
	}

	err = coll.FindOneAndUpdate(context.Background(), filter, update, opt).Decode(&result)
	if err != nil {
		return 0, err
	}

	return result.NextID - int64(count) + 1, nil
}

// InsertOneWithAutoID 插入单个文档，自动生成 ID
func (m *Mongo) InsertOneWithAutoID(ctx context.Context, collectionName string, doc AutoIDModel) (*mongo.InsertOneResult, error) {
	if doc.GetID() == 0 {
		doc.SetID(m.GetIdGenerator(collectionName))
	}
	return m.Collection(collectionName).InsertOne(ctx, doc)
}

// InsertManyWithAutoID 批量插入文档，自动生成 ID
func (m *Mongo) InsertManyWithAutoID(ctx context.Context, collectionName string, docs []AutoIDModel) (*mongo.InsertManyResult, error) {
	if len(docs) == 0 {
		return nil, nil
	}

	startID, err := m.GetBatchIdGenerator(collectionName, len(docs))
	if err != nil {
		return nil, err
	}

	interfaceDocs := make([]interface{}, len(docs))
	for i, doc := range docs {
		if doc.GetID() == 0 {
			doc.SetID(startID + int64(i))
		}
		interfaceDocs[i] = doc
	}

	return m.Collection(collectionName).InsertMany(ctx, interfaceDocs)
}

// MapStr 兼容 map[string]interface{}
type MapStr map[string]interface{}

// Filter 兼容 Filter
type Filter interface{}

// Querier 兼容 Querier
type Querier[T any] interface {
	FindOne(ctx context.Context) (*T, error)
	FindMany(ctx context.Context) ([]*T, error)
}
