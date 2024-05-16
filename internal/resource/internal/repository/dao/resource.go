package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const ResourceCollection = "c_resources"

type ResourceDAO interface {
	CreateResource(ctx context.Context, resource Resource) (int64, error)
	FindResourceById(ctx context.Context, fields []string, id int64) (Resource, error)
	ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]Resource, error)
	Count(ctx context.Context, modelUid string) (int64, error)
	ListResourcesByIds(ctx context.Context, fields []string, ids []int64) ([]Resource, error)
	ListExcludeResourceByids(ctx context.Context, fields []string, modelUid string, offset, limit int64, ids []int64) ([]Resource, error)
	DeleteResource(ctx context.Context, id int64) (int64, error)
}

type resourceDAO struct {
	db *mongox.Mongo
}

func NewResourceDAO(db *mongox.Mongo) ResourceDAO {
	return &resourceDAO{
		db: db,
	}
}

func (dao *resourceDAO) CreateResource(ctx context.Context, r Resource) (int64, error) {
	now := time.Now()
	r.Ctime, r.Utime = now.UnixMilli(), now.UnixMilli()
	r.ID = dao.db.GetIdGenerator(ResourceCollection)
	col := dao.db.Collection(ResourceCollection)

	_, err := col.InsertOne(ctx, r)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return r.ID, nil
}

func (dao *resourceDAO) FindResourceById(ctx context.Context, fields []string, id int64) (Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": id}
	projection := make(map[string]int, len(fields))
	for _, v := range fields {
		projection[v] = 1
	}
	projection["_id"] = 0
	projection["id"] = 1
	projection["name"] = 1
	opts := &options.FindOneOptions{
		Projection: projection,
	}

	var result Resource
	if err := col.FindOne(ctx, filter, opts).Decode(&result); err != nil {
		return Resource{}, fmt.Errorf("解码错误: %w", err)
	}
	return result, nil
}

func (dao *resourceDAO) ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"model_uid": modelUid}
	projection := make(map[string]int, len(fields))
	for _, v := range fields {
		projection[v] = 1
	}
	projection["_id"] = 0
	projection["id"] = 1
	projection["name"] = 1

	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
	}

	cursor, err := col.Find(ctx, filter, opts)
	var result []Resource
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *resourceDAO) Count(ctx context.Context, modelUid string) (int64, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"model_uid": modelUid}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *resourceDAO) ListResourcesByIds(ctx context.Context, fields []string, ids []int64) ([]Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": bson.M{"$in": ids}}
	projection := make(map[string]int, len(fields))
	for _, v := range fields {
		projection[v] = 1
	}
	projection["_id"] = 0
	projection["id"] = 1
	projection["name"] = 1
	projection["model_uid"] = 1
	opts := &options.FindOptions{
		Projection: projection,
	}

	cursor, err := col.Find(ctx, filter, opts)
	var result []Resource
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *resourceDAO) ListExcludeResourceByids(ctx context.Context, fields []string, modelUid string, offset, limit int64, ids []int64) ([]Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"model_uid": modelUid}

	if len(ids) > 0 {
		filter["id"] = bson.M{
			"$nin": ids,
		}
	}
	projection := make(map[string]int, len(fields))
	for _, v := range fields {
		projection[v] = 1
	}
	projection["_id"] = 0
	projection["id"] = 1
	projection["name"] = 1

	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
	}

	cursor, err := col.Find(ctx, filter, opts)
	var result []Resource
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *resourceDAO) DeleteResource(ctx context.Context, id int64) (int64, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": id}

	result, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

type Resource struct {
	ID       int64         `bson:"id"`
	Name     string        `bson:"name"`
	ModelUID string        `bson:"model_uid"`
	Data     mongox.MapStr `bson:",inline"`
	Ctime    int64         `bson:"ctime"`
	Utime    int64         `bson:"utime"`
}
