package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const AttributeGroupCollection = "c_attribute_group"

type AttributeGroupDAO interface {
	CreateAttributeGroup(ctx context.Context, req AttributeGroup) (int64, error)
	ListAttributeGroup(ctx context.Context, modelUid string) ([]AttributeGroup, error)
}

type attributeGroupDAO struct {
	db *mongox.Mongo
}

func NewAttributeGroupDAO(db *mongox.Mongo) AttributeGroupDAO {
	return &attributeGroupDAO{
		db: db,
	}
}

func (dao *attributeGroupDAO) CreateAttributeGroup(ctx context.Context, req AttributeGroup) (int64, error) {
	req.Id = dao.db.GetIdGenerator(AttributeGroupCollection)
	col := dao.db.Collection(AttributeCollection)
	now := time.Now()
	req.Ctime, req.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return req.Id, nil
}

func (dao *attributeGroupDAO) ListAttributeGroup(ctx context.Context, modelUid string) ([]AttributeGroup, error) {
	col := dao.db.Collection(AttributeGroupCollection)
	filter := bson.M{"model_uid": modelUid}
	opts := &options.FindOptions{}

	cursor, err := col.Find(ctx, filter, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []AttributeGroup
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

type AttributeGroup struct {
	Id       int64  `bson:"id"`
	Name     string `bson:"name"`
	ModelUid string `bson:"model_uid"`
	Index    string `bson:"index"`
	Ctime    int64  `bson:"ctime"`
	Utime    int64  `bson:"utime"`
}
