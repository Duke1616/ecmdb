package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type AttributeDAO interface {
	CreateAttribute(ctx context.Context, ab Attribute) (int64, error)
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
