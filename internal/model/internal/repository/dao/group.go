package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type ModelGroupDAO interface {
	CreateModelGroup(ctx context.Context, mg ModelGroup) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]ModelGroup, error)
	Count(ctx context.Context) (int64, error)
}

func NewModelGroupDAO(db *mongox.Mongo) ModelGroupDAO {
	return &groupDAO{
		db: db,
	}
}

type groupDAO struct {
	db *mongox.Mongo
}

func (dao *groupDAO) List(ctx context.Context, offset, limit int64) ([]ModelGroup, error) {
	col := dao.db.Collection(ModelGroupCollection)
	filer := bson.M{}
	opt := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: 1}},
		Limit: &limit,
		Skip:  &offset,
	}
	cursor, err := col.Find(ctx, filer, opt)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []ModelGroup
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *groupDAO) Count(ctx context.Context) (int64, error) {
	col := dao.db.Collection(ModelGroupCollection)
	filter := bson.M{}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *groupDAO) CreateModelGroup(ctx context.Context, mg ModelGroup) (int64, error) {
	now := time.Now()
	mg.Ctime, mg.Utime = now.UnixMilli(), now.UnixMilli()
	mg.Id = dao.db.GetIdGenerator(ModelGroupCollection)
	col := dao.db.Collection(ModelGroupCollection)

	_, err := col.InsertOne(ctx, mg)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return mg.Id, nil
}
