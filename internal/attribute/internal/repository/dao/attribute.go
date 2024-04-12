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
	SearchAttributeByModelIdentifies(ctx context.Context, identifies string) ([]*Attribute, error)
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

func (dao *attributeDAO) SearchAttributeByModelIdentifies(ctx context.Context, identifies string) ([]*Attribute, error) {
	col := dao.db.Collection(AttributeCollection)

	filer := bson.M{"model_identifies": identifies}
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

type Attribute struct {
	Id              int64  `bson:"id"`
	ModelIdentifies string `bson:"model_identifies"`
	Name            string `bson:"name"`
	Identifies      string `bson:"identifies"`
	FieldType       string `bson:"field_type"`
	Required        bool   `bson:"required"`
	Ctime           int64  `bson:"ctime"`
	Utime           int64  `bson:"utime"`
}
