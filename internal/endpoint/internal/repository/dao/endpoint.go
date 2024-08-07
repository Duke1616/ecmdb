package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"time"
)

const EndpointCollection = "c_endpoint"

type EndpointDAO interface {
	CreateEndpoint(ctx context.Context, t Endpoint) (int64, error)
}

type endpointDAO struct {
	db *mongox.Mongo
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
