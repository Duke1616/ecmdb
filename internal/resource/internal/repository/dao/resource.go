package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const ResourceCollection = "c_resources"

type ResourceDAO interface {
	CreateResource(ctx context.Context, resource Resource) (int64, error)
	FindResourceById(ctx context.Context, projection map[string]int, id int64) (Resource, error)
	ListResource(ctx context.Context, projection map[string]int, modelUid string, offset, limit int64) ([]Resource, error)

	ListResourcesByIds(ctx context.Context, projection map[string]int, ids []int64) ([]*Resource, error)

	FindResource(ctx context.Context, id int64) (Resource, error)

	ListExcludeResource(ctx context.Context, projection map[string]int, modelUid string, offset, limit int64, ids []int64) ([]Resource, error)
}

type resourceDAO struct {
	db *mongox.Mongo
}

func NewResourceDAO(db *mongox.Mongo) ResourceDAO {
	return &resourceDAO{
		db: db,
	}
}

func (dao *resourceDAO) CreateResource(ctx context.Context, r Resource) (int64, error) {
	r.ID = dao.db.GetIdGenerator(ResourceCollection)
	col := dao.db.Collection(ResourceCollection)

	_, err := col.InsertMany(ctx, []interface{}{r})

	if err != nil {
		return 0, err
	}

	return r.ID, nil
}

func (dao *resourceDAO) FindResourceById(ctx context.Context, projection map[string]int, id int64) (Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": id}
	projection["_id"] = 0
	projection["id"] = 1
	projection["name"] = 1

	opts := &options.FindOptions{
		Projection: projection,
	}

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return Resource{}, err
	}

	var result Resource
	for cursor.Next(ctx) {
		if err = cursor.Decode(&result); err != nil {
			return Resource{}, err
		}
	}

	return result, nil
}

func (dao *resourceDAO) ListResource(ctx context.Context, projection map[string]int, modelUid string, offset, limit int64) ([]Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"model_uid": modelUid}
	projection["_id"] = 0
	projection["id"] = 1
	projection["name"] = 1

	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
	}

	cursor, err := col.Find(ctx, filter, opts)
	var result []Resource
	for cursor.Next(ctx) {
		var rs Resource
		if err = cursor.Decode(&rs); err != nil {
			return nil, err
		}
		result = append(result, rs)
	}

	return result, nil
}

func (dao *resourceDAO) ListResourcesByIds(ctx context.Context, projection map[string]int, ids []int64) ([]*Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": bson.M{"$in": ids}}
	projection["_id"] = 0
	projection["id"] = 1
	projection["name"] = 1
	opts := &options.FindOptions{
		Projection: projection,
	}

	cursor, err := col.Find(ctx, filter, opts)

	result := make([]*Resource, 0)
	for cursor.Next(ctx) {
		if err = cursor.Decode(&result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (dao *resourceDAO) FindResource(ctx context.Context, id int64) (Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": id}
	opts := &options.FindOptions{}

	cursor, err := col.Find(ctx, filter, opts)

	var result Resource
	for cursor.Next(ctx) {
		if err = cursor.Decode(&result); err != nil {
			return Resource{}, err
		}
	}

	return result, nil
}

func (dao *resourceDAO) ListExcludeResource(ctx context.Context, projection map[string]int, modelUid string, offset, limit int64, ids []int64) ([]Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"model_uid": modelUid}

	if len(ids) > 0 {
		filter["id"] = bson.M{
			"$nin": ids,
		}
	}

	projection["_id"] = 0
	projection["id"] = 1
	projection["name"] = 1

	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
	}

	cursor, err := col.Find(ctx, filter, opts)
	var result []Resource
	for cursor.Next(ctx) {
		var rs Resource
		if err = cursor.Decode(&rs); err != nil {
			return nil, err
		}
		result = append(result, rs)
	}

	return result, nil
}

type Resource struct {
	ID       int64         `bson:"id"`
	Name     string        `bson:"name"`
	ModelUID string        `bson:"model_uid"`
	Data     mongox.MapStr `bson:",inline"`
	Ctime    int64         `bson:"ctime"`
	Utime    int64         `bson:"utime"`
}
