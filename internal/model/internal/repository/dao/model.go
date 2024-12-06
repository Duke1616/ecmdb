package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log/slog"
	"time"
)

type ModelDAO interface {
	Create(ctx context.Context, m Model) (int64, error)
	GetById(ctx context.Context, id int64) (Model, error)
	GetByUids(ctx context.Context, uids []string) ([]Model, error)
	List(ctx context.Context, offset, limit int64) ([]Model, error)
	Count(ctx context.Context) (int64, error)
	ListAll(ctx context.Context) ([]Model, error)

	ListByGroupIds(ctx context.Context, mgids []int64) ([]Model, error)
	DeleteById(ctx context.Context, id int64) (int64, error)
	DeleteByUid(ctx context.Context, modelUid string) (int64, error)
}

func NewModelDAO(db *mongox.Mongo) ModelDAO {
	return &modelDAO{
		db: db,
	}
}

type modelDAO struct {
	db *mongox.Mongo
}

func (dao *modelDAO) GetByUids(ctx context.Context, uids []string) ([]Model, error) {
	col := dao.db.Collection(ModelCollection)
	filter := bson.M{"uid": bson.M{"$in": uids}}
	cursor, err := col.Find(ctx, filter)
	var result []Model
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *modelDAO) ListByGroupIds(ctx context.Context, mgids []int64) ([]Model, error) {
	col := dao.db.Collection(ModelCollection)
	filter := bson.M{}

	if len(mgids) <= 0 {
		slog.Warn("没有匹配的数据, 模型组为空")
		return nil, nil
	}

	filter["model_group_id"] = bson.M{
		"$in": mgids,
	}

	opts := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: 1}},
	}

	cursor, err := col.Find(ctx, filter, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []Model
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *modelDAO) Create(ctx context.Context, m Model) (int64, error) {
	now := time.Now()
	m.Ctime, m.Utime = now.UnixMilli(), now.UnixMilli()
	m.Id = dao.db.GetIdGenerator(ModelCollection)
	col := dao.db.Collection(ModelCollection)

	_, err := col.InsertOne(ctx, m)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return m.Id, nil
}

func (dao *modelDAO) GetById(ctx context.Context, id int64) (Model, error) {
	col := dao.db.Collection(ModelCollection)
	filter := bson.M{"id": id}

	var m Model
	if err := col.FindOne(ctx, filter).Decode(&m); err != nil {
		return Model{}, fmt.Errorf("解码错误: %w", err)
	}

	return m, nil
}

func (dao *modelDAO) ListAll(ctx context.Context) ([]Model, error) {
	col := dao.db.Collection(ModelCollection)
	filter := bson.M{}
	opt := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}
	cursor, err := col.Find(ctx, filter, opt)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []Model
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *modelDAO) List(ctx context.Context, offset, limit int64) ([]Model, error) {
	col := dao.db.Collection(ModelCollection)
	filter := bson.M{}
	opt := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}
	cursor, err := col.Find(ctx, filter, opt)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []Model
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *modelDAO) Count(ctx context.Context) (int64, error) {
	col := dao.db.Collection(ModelCollection)
	filter := bson.M{}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *modelDAO) DeleteById(ctx context.Context, id int64) (int64, error) {
	col := dao.db.Collection(ModelCollection)
	filter := bson.M{"id": id}

	result, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *modelDAO) DeleteByUid(ctx context.Context, modelUid string) (int64, error) {
	col := dao.db.Collection(ModelCollection)
	filter := bson.M{"uid": modelUid}

	result, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}
