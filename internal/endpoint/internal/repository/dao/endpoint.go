package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const EndpointCollection = "c_endpoint"

type EndpointDAO interface {
	CreateEndpoint(ctx context.Context, t Endpoint) (int64, error)
	CreateMutilEndpoint(ctx context.Context, req []Endpoint) (int64, error)
	ListEndpoint(ctx context.Context, offset, limit int64, path string) ([]Endpoint, error)
	Count(ctx context.Context, path string) (int64, error)
}

type endpointDAO struct {
	db *mongox.Mongo
}

func (dao *endpointDAO) CreateMutilEndpoint(ctx context.Context, req []Endpoint) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (dao *endpointDAO) ListEndpoint(ctx context.Context, offset, limit int64, path string) ([]Endpoint, error) {
	col := dao.db.Collection(EndpointCollection)
	filter := bson.M{}
	if path != "" {
		filter = bson.M{"$text": bson.M{"$search": path}}
	}
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	cursor, err := col.Find(ctx, filter, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []Endpoint
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *endpointDAO) Count(ctx context.Context, path string) (int64, error) {
	col := dao.db.Collection(EndpointCollection)
	filter := bson.M{}
	if path != "" {
		filter = bson.M{"$text": bson.M{"$search": path}}
	}
	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *endpointDAO) CreateEndpoint(ctx context.Context, e Endpoint) (int64, error) {
	e.Id = dao.db.GetIdGenerator(EndpointCollection)
	col := dao.db.Collection(EndpointCollection)
	now := time.Now()
	e.Ctime, e.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, e)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return e.Id, nil
}

func NewEndpointDAO(db *mongox.Mongo) EndpointDAO {
	return &endpointDAO{
		db: db,
	}
}

type Endpoint struct {
	Id           int64  `bson:"id"`
	Path         string `bson:"path"`
	Method       string `bson:"method"`
	Resource     string `bson:"resource"`
	Desc         string `bson:"desc"`
	IsAuth       bool   `bson:"is_auth"`
	IsAudit      bool   `bson:"is_audit"`
	IsPermission bool   `bson:"is_permission"`
	Ctime        int64  `bson:"ctime"`
	Utime        int64  `bson:"utime"`
}
