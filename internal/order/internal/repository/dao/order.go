package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

const (
	OrderCollection = "c_order"
)

type OrderDAO interface {
	CreateOrder(ctx context.Context, r Order) (int64, error)
	RegisterProcessInstanceId(ctx context.Context, id int64, instanceId int, status uint8) error
	ListOrderByProcessEngineIds(ctx context.Context, engineIds []int) ([]Order, error)
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

func (dao *orderDAO) RegisterProcessInstanceId(ctx context.Context, id int64, instanceId int, status uint8) error {
	col := dao.db.Collection(OrderCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"process_instance_id": instanceId,
			"status":              status,
			"utime":               time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": id}
	_, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("修改文档操作: %w", err)
	}

	return nil
}

func (dao *orderDAO) ListOrderByProcessEngineIds(ctx context.Context, engineIds []int) ([]Order, error) {
	col := dao.db.Collection(OrderCollection)
	filter := bson.M{"process_engine_id": bson.M{"$in": engineIds}}

	cursor, err := col.Find(ctx, filter)
	var result []Order
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

type Order struct {
	Id                int64                  `bson:"id"`
	TemplateId        int64                  `bson:"template_id"`
	WorkflowId        int64                  `bson:"workflow_id"`
	ProcessInstanceId int                    `bson:"process_instance_id"`
	CreateBy          string                 `bson:"create_by"`
	Data              map[string]interface{} `bson:"data"`
	Status            uint8                  `bson:"status"`
	Ctime             int64                  `bson:"ctime"`
	Utime             int64                  `bson:"utime"`
}
