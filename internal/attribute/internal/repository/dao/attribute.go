package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const AttributeCollection = "c_attribute"

type AttributeDAO interface {
	CreateAttribute(ctx context.Context, ab Attribute) (int64, error)
	SearchAttributeByModelUID(ctx context.Context, modelUid string) ([]*Attribute, error)

	ListAttribute(ctx context.Context, modelUID string) ([]Attribute, error)
}

type attributeDAO struct {
	db *mongox.Mongo
}

func NewAttributeDAO(client *mongo.Client) AttributeDAO {
	return &attributeDAO{
		db: mongox.NewMongo(client),
	}
}

func (dao *attributeDAO) CreateAttribute(ctx context.Context, attribute Attribute) (int64, error) {
	now := time.Now()
	attribute.Ctime, attribute.Utime = now.UnixMilli(), now.UnixMilli()
	attribute.Id = dao.db.GetIdGenerator(AttributeCollection)
	col := dao.db.Collection(AttributeCollection)

	_, err := col.InsertMany(ctx, []interface{}{attribute})

	if err != nil {
		return 0, err
	}

	return attribute.Id, nil
}

func (dao *attributeDAO) SearchAttributeByModelUID(ctx context.Context, modelUid string) ([]*Attribute, error) {
	col := dao.db.Collection(AttributeCollection)

	filer := bson.M{"model_uid": modelUid}
	opt := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	resp, err := col.Find(ctx, filer, opt)
	var set []*Attribute
	for resp.Next(ctx) {
		ins := &Attribute{}
		if err = resp.Decode(ins); err != nil {
			return nil, err
		}
		set = append(set, ins)
	}

	return set, nil
}

func (dao *attributeDAO) ListAttribute(ctx context.Context, modelUID string) ([]Attribute, error) {
	col := dao.db.Collection(AttributeCollection)
	filter := bson.M{"model_uid": modelUID}

	opts := &options.FindOptions{}

	cursor, err := col.Find(ctx, filter, opts)

	result := make([]Attribute, 0)
	for cursor.Next(ctx) {
		var at Attribute
		if err = cursor.Decode(&at); err != nil {
			return nil, err
		}
		result = append(result, at)
	}

	return result, nil
}

type Attribute struct {
	Id        int64  `bson:"id"`
	ModelUID  string `bson:"model_uid"`  // 模型唯一标识
	Name      string `bson:"name"`       // 字段名称
	UID       string `bson:"uid"`        // 字段唯一标识、英文标识
	FieldType string `bson:"field_type"` // 字段类型
	Required  bool   `bson:"required"`   // 是否为必传
	Display   bool   `bson:"display"`    // 是否前端展示
	Index     int64  `bson:"index"`      // 字段前端展示顺序
	Ctime     int64  `bson:"ctime"`
	Utime     int64  `bson:"utime"`
}
