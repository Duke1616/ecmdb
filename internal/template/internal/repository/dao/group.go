package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TemplateGroupDAO interface {
	Create(ctx context.Context, t TemplateGroup) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]TemplateGroup, error)
	Count(ctx context.Context) (int64, error)
	ListByIds(ctx context.Context, ids []int64) ([]TemplateGroup, error)
}

func NewTemplateGroupDAO(db *mongox.Mongo) TemplateGroupDAO {
	return &templateGroupDAO{
		db: db,
	}
}

type templateGroupDAO struct {
	db *mongox.Mongo
}

func (dao *templateGroupDAO) Create(ctx context.Context, t TemplateGroup) (int64, error) {
	t.Id = dao.db.GetIdGenerator(TemplateGroupCollection)
	col := dao.db.Collection(TemplateGroupCollection)
	now := time.Now()
	t.Ctime, t.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, t)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return t.Id, nil
}

func (dao *templateGroupDAO) List(ctx context.Context, offset, limit int64) ([]TemplateGroup, error) {
	col := dao.db.Collection(TemplateGroupCollection)
	filter := bson.M{}
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

	var result []TemplateGroup
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *templateGroupDAO) Count(ctx context.Context) (int64, error) {
	col := dao.db.Collection(TemplateGroupCollection)
	filer := bson.M{}

	count, err := col.CountDocuments(ctx, filer)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *templateGroupDAO) ListByIds(ctx context.Context, ids []int64) ([]TemplateGroup, error) {
	col := dao.db.Collection(TemplateGroupCollection)
	filter := bson.M{"id": bson.M{"$in": ids}}

	cursor, err := col.Find(ctx, filter)
	var result []TemplateGroup
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}
