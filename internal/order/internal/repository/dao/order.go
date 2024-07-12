package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"time"
)

const (
	OrderCollection = "c_order"
)

type OrderDAO interface {
	CreateOrder(ctx context.Context, r Order) (int64, error)
}

func NewOrderDAO(db *mongox.Mongo) OrderDAO {
	return &orderDAO{
		db: db,
	}
}

type orderDAO struct {
	db *mongox.Mongo
}

func (dao *orderDAO) CreateOrder(ctx context.Context, r Order) (int64, error) {
	now := time.Now()
	r.Ctime, r.Utime = now.UnixMilli(), now.UnixMilli()
	r.Id = dao.db.GetIdGenerator(OrderCollection)
	col := dao.db.Collection(OrderCollection)

	_, err := col.InsertOne(ctx, r)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return r.Id, nil
}

type Order struct {
	Id         int64                  `bson:"id"`
	TemplateId int64                  `bson:"template_idp"`
	FlowId     int64                  `bson:"flow_id"`
	CreateBy   string                 `bson:"create_by"`
	Data       map[string]interface{} `bson:"data"`
	Status     uint8                  `bson:"status"`
	Ctime      int64                  `bson:"ctime"`
	Utime      int64                  `bson:"utime"`
}
