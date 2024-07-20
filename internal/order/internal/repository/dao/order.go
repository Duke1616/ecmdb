package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	OrderCollection = "c_order"
)

type OrderDAO interface {
	CreateOrder(ctx context.Context, r Order) (int64, error)
	DetailByProcessInstId(ctx context.Context, instanceId int) (Order, error)
	RegisterProcessInstanceId(ctx context.Context, id int64, instanceId int, status uint8) error
	ListOrderByProcessInstanceIds(ctx context.Context, instanceIds []int) ([]Order, error)
	UpdateStatusByInstanceId(ctx context.Context, instanceId int, status uint8) error
	ListOrder(ctx context.Context, userId string, status uint8, offset, limit int64) ([]Order, error)
	CountOrder(ctx context.Context, userId string, status uint8) (int64, error)
}

func NewOrderDAO(db *mongox.Mongo) OrderDAO {
	return &orderDAO{
		db: db,
	}
}

type orderDAO struct {
	db *mongox.Mongo
}

func (dao *orderDAO) UpdateStatusByInstanceId(ctx context.Context, instanceId int, status uint8) error {
	col := dao.db.Collection(OrderCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"status": status,
			"utime":  time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"process_instance_id": instanceId}
	_, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("修改文档操作: %w", err)
	}

	return nil
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

func (dao *orderDAO) DetailByProcessInstId(ctx context.Context, instanceId int) (Order, error) {
	col := dao.db.Collection(OrderCollection)
	filter := bson.M{"process_instance_id": instanceId}

	var result Order
	if err := col.FindOne(ctx, filter).Decode(&result); err != nil {
		return Order{}, fmt.Errorf("解码错误，%w", err)
	}

	return result, nil
}

func (dao *orderDAO) ListOrderByProcessInstanceIds(ctx context.Context, instanceIds []int) ([]Order, error) {
	col := dao.db.Collection(OrderCollection)
	filter := bson.M{"process_instance_id": bson.M{"$in": instanceIds}}

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

func (dao *orderDAO) ListOrder(ctx context.Context, userId string, status uint8, offset, limit int64) ([]Order, error) {
	col := dao.db.Collection(OrderCollection)
	filter := bson.M{}
	if userId != "" {
		filter = bson.M{"create_by": userId}
	}

	if status != 0 {
		filter = bson.M{"status": status}
	}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}
	cursor, err := col.Find(ctx, filter, opts)
	var result []Order
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *orderDAO) CountOrder(ctx context.Context, userId string, status uint8) (int64, error) {
	col := dao.db.Collection(OrderCollection)
	filter := bson.M{}
	if userId != "" {
		filter = bson.M{"create_by": userId}
	}

	if status != 0 {
		filter = bson.M{"status": status}
	}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

type Order struct {
	Id                int64                  `bson:"id"`
	TemplateId        int64                  `bson:"template_id"`
	TemplateName      string                 `bson:"template_name"`
	WorkflowId        int64                  `bson:"workflow_id"`
	ProcessInstanceId int                    `bson:"process_instance_id"`
	CreateBy          string                 `bson:"create_by"`
	Data              map[string]interface{} `bson:"data"`
	Status            uint8                  `bson:"status"`
	Ctime             int64                  `bson:"ctime"`
	Utime             int64                  `bson:"utime"`
}
