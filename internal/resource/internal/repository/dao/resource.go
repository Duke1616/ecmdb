package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type ResourceDAO interface {
	CreateResource(ctx context.Context, data mongox.MapStr, ab Resource) (int64, error)
	FindResourceById(ctx context.Context, dmAttr domain.DetailResource) ([]mongox.MapStr, error)
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

func (dao *resourceDAO) FindResourceById(ctx context.Context, dmAttr domain.DetailResource) ([]mongox.MapStr, error) {
	col := dao.db.Collection("c_resources")
	filter := bson.M{"id": dmAttr.ID}
	dmAttr.Projection["id"] = 1

	opts := &options.FindOptions{
		Projection: dmAttr.Projection,
	}

	resources := make([]mongox.MapStr, 0)
	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	err = cursor.All(ctx, &resources)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

type Resource struct {
	ModelIdentifies string
}
