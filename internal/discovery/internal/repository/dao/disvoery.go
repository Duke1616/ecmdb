package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/ekit/slice"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// ErrDataNotFound 通用的数据没找到
var ErrDataNotFound = mongo.ErrNoDocuments

const (
	DiscoveryCollection = "c_discovery"
)

type DiscoveryDAO interface {
	Create(ctx context.Context, req Discovery) (int64, error)
	Update(ctx context.Context, req Discovery) (int64, error)
	Delete(ctx context.Context, id int64) (int64, error)
	ListByTemplateId(ctx context.Context, offset, limit int64, templateId int64) ([]Discovery, error)
	CountByTemplateId(ctx context.Context, templateId int64) (int64, error)
	Sync(ctx context.Context, templateId int64, docs []Discovery) (int64, error)
}

type discoveryDao struct {
	db *mongox.Mongo
}

func (dao *discoveryDao) Sync(ctx context.Context, templateId int64, docs []Discovery) (int64, error) {
	col := dao.db.Collection(DiscoveryCollection)
	upsert := true
	now := time.Now().UnixMilli()

	operations := make([]mongo.WriteModel, len(docs))
	operations = slice.Map(docs, func(idx int, src Discovery) mongo.WriteModel {
		return &mongo.UpdateOneModel{
			Filter: bson.D{
				{"template_id", templateId},
				{"field", src.Field},
				{"value", src.Value},
			},
			// 可以仅使用 $setOnInsert
			Update: bson.D{
				{"$setOnInsert", bson.D{
					{"id", dao.db.GetIdGenerator(DiscoveryCollection)},
					{"ctime", now},
				}},
				{"$set", bson.D{
					{"field", src.Field},
					{"value", src.Value},
					{"utime", now},
				}},
			},
			Upsert: &upsert,
		}
	})

	result, err := col.BulkWrite(ctx, operations)
	if err != nil {
		return 0, fmt.Errorf("BulkFindOrInsert failed: %w", err)
	}

	return result.UpsertedCount, nil // 返回插入的文档数
}

func (dao *discoveryDao) Delete(ctx context.Context, id int64) (int64, error) {
	col := dao.db.Collection(DiscoveryCollection)
	filter := bson.M{"id": id}

	result, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *discoveryDao) Create(ctx context.Context, req Discovery) (int64, error) {
	req.Id = dao.db.GetIdGenerator(DiscoveryCollection)
	col := dao.db.Collection(DiscoveryCollection)
	now := time.Now()
	req.Ctime, req.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return req.Id, nil
}

func (dao *discoveryDao) Update(ctx context.Context, req Discovery) (int64, error) {
	col := dao.db.Collection(DiscoveryCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"runner_id": req.RunnerId,
			"field":     req.Field,
			"value":     req.Value,
			"utime":     time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": req.Id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *discoveryDao) ListByTemplateId(ctx context.Context, offset, limit int64, templateId int64) ([]Discovery, error) {
	col := dao.db.Collection(DiscoveryCollection)
	filter := bson.M{}
	filter["template_id"] = bson.M{"$eq": templateId}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	cursor, err := col.Find(ctx, filter, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []Discovery
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *discoveryDao) CountByTemplateId(ctx context.Context, templateId int64) (int64, error) {
	col := dao.db.Collection(DiscoveryCollection)
	filter := bson.M{}
	filter["template_id"] = bson.M{"$eq": templateId}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func NewDiscoveryDAO(db *mongox.Mongo) DiscoveryDAO {
	return &discoveryDao{
		db: db,
	}
}

type Discovery struct {
	Id         int64  `bson:"id"`
	TemplateId int64  `bson:"template_id"`
	RunnerId   int64  `bson:"runner_id"`
	Field      string `bson:"field"`
	Value      string `bson:"value"`
	Ctime      int64  `bson:"ctime"`
	Utime      int64  `bson:"utime"`
}
