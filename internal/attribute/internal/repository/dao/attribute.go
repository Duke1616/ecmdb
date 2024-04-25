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
	now := time.Now()
	attr.Ctime, attr.Utime = now.UnixMilli(), now.UnixMilli()
	attr.Id = dao.db.GetIdGenerator(AttributeCollection)
	col := dao.db.Collection(AttributeCollection)

	_, err := col.InsertOne(ctx, attr)

	if err != nil {
		return 0, err
	}

	return attr.Id, nil
}

func (dao *attributeDAO) SearchAttributeByModelUID(ctx context.Context, modelUid string) ([]Attribute, error) {
	col := dao.db.Collection(AttributeCollection)

	filer := bson.M{"model_uid": modelUid}
	opt := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	resp, err := col.Find(ctx, filer, opt)
	var set []Attribute
	for resp.Next(ctx) {
		var ins Attribute
		if err = resp.Decode(&ins); err != nil {
			return nil, err
		}
		set = append(set, ins)
	}

	return set, nil
}

func (dao *attributeDAO) ListAttribute(ctx context.Context, modelUid string) ([]Attribute, error) {
	col := dao.db.Collection(AttributeCollection)
	filter := bson.M{"model_uid": modelUid}
	opts := &options.FindOptions{}

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	result := make([]Attribute, 0)
	for cursor.Next(ctx) {
		var attr Attribute
		if err = cursor.Decode(&attr); err != nil {
			return nil, err
		}
		result = append(result, attr)
	}

	return result, nil
}

func (dao *attributeDAO) Count(ctx context.Context, modelUid string) (int64, error) {
	col := dao.db.Collection(AttributeCollection)
	filter := bson.M{"model_uid": modelUid}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
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
