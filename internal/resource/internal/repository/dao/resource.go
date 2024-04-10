package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type ResourceDAO interface {
	CreateResource(ctx context.Context, ab Resource) (int64, error)
}

type resourceDAO struct {
	db *mongo.Database
}

func NewResourceDAO(client *mongo.Client) ResourceDAO {
	return &resourceDAO{
		db: client.Database("cmdb"),
	}
}

func (dao *resourceDAO) CreateResource(ctx context.Context, resource Resource) (int64, error) {
	now := time.Now()
	resource.Ctime, resource.Utime = now.UnixMilli(), now.UnixMilli()
	resource.Id = mongox.GetDataID(dao.db, "c_resources")

	col := dao.db.Collection("c_resources")
	_, err := col.InsertMany(ctx, []interface{}{resource})

	if err != nil {
		return 0, err
	}

	return resource.Id, nil
}

type Resource struct {
	Id      int64                  `bson:"id"`
	ModelID int64                  `bson:"model_id"`
	Ctime   int64                  `bson:"ctime"`
	Utime   int64                  `bson:"utime"`
	Data    map[string]interface{} `bson:"-"`
}
