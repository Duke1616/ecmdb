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

const ResourceCollection = "c_resources"

type ResourceDAO interface {
	CreateResource(ctx context.Context, resource Resource) (int64, error)
	FindResourceById(ctx context.Context, dmAttr domain.DetailResource) ([]mongox.MapStr, error)

	ListResourcesByIds(ctx context.Context, projection map[string]int, ids []int64) ([]*Resource, error)
}

type resourceDAO struct {
	db *mongox.Mongo
}

func NewResourceDAO(client *mongo.Client) ResourceDAO {
	return &resourceDAO{
		db: mongox.NewMongo(client),
	}
}

func (dao *resourceDAO) CreateResource(ctx context.Context, r Resource) (int64, error) {
	now := time.Now()
	r.Ctime, r.Utime = now.UnixMilli(), now.UnixMilli()
	r.ID = dao.db.GetIdGenerator(ResourceCollection)
	col := dao.db.Collection(ResourceCollection)

	_, err := col.InsertMany(ctx, []interface{}{r})

	if err != nil {
		return 0, err
	}

	return r.ID, nil
}

func (dao *resourceDAO) FindResourceById(ctx context.Context, dmAttr domain.DetailResource) ([]mongox.MapStr, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": dmAttr.ID}
	dmAttr.Projection["_id"] = 0
	dmAttr.Projection["id"] = 1
	dmAttr.Projection["name"] = 1

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

func (dao *resourceDAO) ListResourcesByIds(ctx context.Context, projection map[string]int, ids []int64) ([]*Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": bson.M{"$in": ids}}

	opts := &options.FindOptions{
		Projection: projection,
	}

	resources := make([]*Resource, 0)
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
	ID       int64         `bson:"id"`
	Name     string        `bson:"name"`
	ModelUID string        `bson:"model_uid"`
	Data     mongox.MapStr `bson:",inline"`
	Ctime    int64         `bson:"ctime"`
	Utime    int64         `bson:"utime"`
}
