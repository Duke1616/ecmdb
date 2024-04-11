package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type ResourceDAO interface {
	CreateResource(ctx context.Context, data mongox.MapStr, ab Resource) (int64, error)
	FindResourceById(ctx context.Context, id int64, modelIdentifies string) (*Resource, error)
}

type resourceDAO struct {
	db *mongo.Database
}

func NewResourceDAO(client *mongo.Client) ResourceDAO {
	return &resourceDAO{
		db: client.Database("cmdb"),
	}
}

func (dao *resourceDAO) CreateResource(ctx context.Context, data mongox.MapStr, resource Resource) (int64, error) {
	now := time.Now()
	id := mongox.GetDataID(dao.db, "c_resources")

	col := dao.db.Collection("c_resources")

	data["id"] = id
	data["model_identifies"] = resource.ModelIdentifies
	data["ctime"] = now.UnixMilli()
	data["utime"] = now.UnixMilli()
	_, err := col.InsertMany(ctx, []interface{}{data})

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (dao *resourceDAO) FindResourceById(ctx context.Context, id int64, modelIdentifies string) (*Resource, error) {
	col := dao.db.Collection("c_resources")
	m := &Resource{}
	filter := bson.M{"id": id}

	if err := col.FindOne(ctx, filter).Decode(m); err != nil {
		return m, err
	}

	fmt.Println(m)
	return m, nil
}

type Resource struct {
	Id              int64
	ModelIdentifies string
	Data            mongox.MapStr
}
