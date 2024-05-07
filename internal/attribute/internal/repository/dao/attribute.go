package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const AttributeCollection = "c_attribute"

type AttributeDAO interface {
	CreateAttribute(ctx context.Context, ab Attribute) (int64, error)
	SearchAttributeByModelUID(ctx context.Context, modelUid string) ([]Attribute, error)

	ListAttribute(ctx context.Context, modelUid string) ([]Attribute, error)
	Count(ctx context.Context, modelUid string) (int64, error)
}

type attributeDAO struct {
	db *mongox.Mongo
}

func NewAttributeDAO(db *mongox.Mongo) AttributeDAO {
	return &attributeDAO{
		db: db,
	}
}

func (dao *attributeDAO) CreateAttribute(ctx context.Context, attr Attribute) (int64, error) {
	attr.Id = dao.db.GetIdGenerator(AttributeCollection)
	col := dao.db.Collection(AttributeCollection)
	now := time.Now()
	attr.Ctime, attr.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, attr)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return attr.Id, nil
}

func (dao *attributeDAO) SearchAttributeByModelUID(ctx context.Context, modelUid string) ([]Attribute, error) {
	col := dao.db.Collection(AttributeCollection)
	filer := bson.M{"model_uid": modelUid}
	opt := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	cursor, err := col.Find(ctx, filer, opt)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []Attribute
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *attributeDAO) ListAttribute(ctx context.Context, modelUid string) ([]Attribute, error) {
	col := dao.db.Collection(AttributeCollection)
	filter := bson.M{"model_uid": modelUid}
	opts := &options.FindOptions{}

	cursor, err := col.Find(ctx, filter, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []Attribute
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *attributeDAO) Count(ctx context.Context, modelUid string) (int64, error) {
	col := dao.db.Collection(AttributeCollection)
	filter := bson.M{"model_uid": modelUid}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

type Attribute struct {
	Id        int64  `bson:"id"`
	ModelUID  string `bson:"model_uid"`  // 模型唯一标识
	Name      string `bson:"name"`       // 字段名称
	FieldName string `bson:"field_name"` // 字段唯一标识、英文标识
	FieldType string `bson:"field_type"` // 字段类型
	Required  bool   `bson:"required"`   // 是否为必传
	Display   bool   `bson:"display"`    // 是否前端展示
	Index     int64  `bson:"index"`      // 字段前端展示顺序
	Ctime     int64  `bson:"ctime"`
	Utime     int64  `bson:"utime"`
}
