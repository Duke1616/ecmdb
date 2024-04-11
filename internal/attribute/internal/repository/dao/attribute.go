package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type AttributeDAO interface {
	CreateAttribute(ctx context.Context, ab Attribute) (int64, error)
	SearchAttributeByIdentifies(ctx context.Context, identifies string) ([]*Attribute, error)
}

type attributeDAO struct {
	db *mongo.Database
}

func NewAttributeDAO(client *mongo.Client) AttributeDAO {
	return &attributeDAO{
		db: client.Database("cmdb"),
	}
}

func (a *attributeDAO) CreateAttribute(ctx context.Context, attribute Attribute) (int64, error) {
	now := time.Now()
	attribute.Ctime, attribute.Utime = now.UnixMilli(), now.UnixMilli()
	attribute.Id = mongox.GetDataID(a.db, "c_attribute")

	col := a.db.Collection("c_attribute")
	_, err := col.InsertMany(ctx, []interface{}{attribute})

	if err != nil {
		return 0, err
	}

	return attribute.Id, nil
}

func (a *attributeDAO) SearchAttributeByIdentifies(ctx context.Context, identifies string) ([]*Attribute, error) {
	col := a.db.Collection("c_attribute")

	filer := bson.M{"identifies": identifies}
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
	Id         int64  `bson:"id"`
	ModelID    int64  `bson:"model_id"`
	Name       string `bson:"name"`
	Identifies string `bson:"identifies"`
	FieldType  string `bson:"field_type"`
	Required   bool   `bson:"required"`
	Ctime      int64  `bson:"ctime"`
	Utime      int64  `bson:"utime"`
}
