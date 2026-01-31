package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	OrderCollection = "c_order"
)

type OrderDAO interface {
	// CreateBizOrder TODO 创建业务工单
	CreateBizOrder(ctx context.Context, order Order) (Order, error)

	// CreateOrder 创建工单
	CreateOrder(ctx context.Context, r Order) (int64, error)
	// DetailByProcessInstId 根据流程实例ID获取工单详情
	DetailByProcessInstId(ctx context.Context, instanceId int) (Order, error)
	// Detail 根据ID获取工单详情
	Detail(ctx context.Context, id int64) (Order, error)
	// RegisterProcessInstanceId 绑定流程实例ID
	RegisterProcessInstanceId(ctx context.Context, id int64, instanceId int, status uint8) error
	// ListOrderByProcessInstanceIds 根据流程实例ID列表获取工单
	ListOrderByProcessInstanceIds(ctx context.Context, instanceIds []int) ([]Order, error)
	// UpdateStatusByInstanceId 根据流程实例ID更新状态
	UpdateStatusByInstanceId(ctx context.Context, instanceId int, status uint8) error
	// ListOrder 获取工单列表
	ListOrder(ctx context.Context, userId string, status []int, offset, limit int64) ([]Order, error)
	// CountOrder 获取工单数量
	CountOrder(ctx context.Context, userId string, status []int) (int64, error)
	// FindByBizIdAndKey 根据 BizID 和 Key 查询工单
	FindByBizIdAndKey(ctx context.Context, bizId int64, key string, status []uint8) (Order, error)
	// MergeOrderData 合并工单数据（原子更新）
	MergeOrderData(ctx context.Context, id int64, data map[string]interface{}) error
}

func (dao *orderDAO) MergeOrderData(ctx context.Context, id int64, data map[string]interface{}) error {
	col := dao.db.Collection(OrderCollection)
	updates := bson.M{
		"utime": time.Now().UnixMilli(),
	}
	for k, v := range data {
		updates["data."+k] = v
	}

	updateDoc := bson.M{
		"$set": updates,
	}
	filter := bson.M{"id": id}
	_, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("修改文档操作: %w", err)
	}

	return nil
}

func NewOrderDAO(db *mongox.Mongo) OrderDAO {
	return &orderDAO{
		db: db,
	}
}

type orderDAO struct {
	db *mongox.Mongo
}

func (dao *orderDAO) CreateBizOrder(ctx context.Context, order Order) (Order, error) {
	now := time.Now()
	order.Ctime, order.Utime = now.UnixMilli(), now.UnixMilli()
	order.Id = dao.db.GetIdGenerator(OrderCollection)
	col := dao.db.Collection(OrderCollection)

	_, err := col.InsertOne(ctx, order)

	if err != nil {
		return order, fmt.Errorf("插入数据错误: %w", err)
	}

	return order, nil
}

func (dao *orderDAO) Detail(ctx context.Context, id int64) (Order, error) {
	col := dao.db.Collection(OrderCollection)
	filter := bson.M{"id": id}

	var result Order
	if err := col.FindOne(ctx, filter).Decode(&result); err != nil {
		return Order{}, fmt.Errorf("解码错误，%w", err)
	}

	return result, nil
}

func (dao *orderDAO) UpdateStatusByInstanceId(ctx context.Context, instanceId int, status uint8) error {
	col := dao.db.Collection(OrderCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"status": status,
			"utime":  time.Now().UnixMilli(),
			"wtime":  time.Now().UnixMilli(),
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

func (dao *orderDAO) ListOrder(ctx context.Context, userId string, status []int, offset, limit int64) ([]Order, error) {
	col := dao.db.Collection(OrderCollection)
	filter := bson.M{}
	if userId != "" {
		filter = bson.M{"create_by": userId}
	}

	if status != nil && len(status) > 0 {
		filter = bson.M{"status": bson.M{"$in": status}}
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

func (dao *orderDAO) CountOrder(ctx context.Context, userId string, status []int) (int64, error) {
	col := dao.db.Collection(OrderCollection)
	filter := bson.M{}
	if userId != "" {
		filter = bson.M{"create_by": userId}
	}

	if status != nil && len(status) > 0 {
		filter = bson.M{"status": bson.M{"$in": status}}
	}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *orderDAO) FindByBizIdAndKey(ctx context.Context, bizId int64, key string, status []uint8) (Order, error) {
	col := dao.db.Collection(OrderCollection)
	filter := bson.M{
		"biz_id": bizId,
		"key":    key,
	}
	if len(status) > 0 {
		filter["status"] = bson.M{"$in": status}
	}

	// 按创建时间倒序，取最新的一个
	opts := &options.FindOneOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	var result Order
	if err := col.FindOne(ctx, filter, opts).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// 没有找到记录，返回空 Order
			return Order{}, nil
		}
		return Order{}, fmt.Errorf("查询工单失败: %w", err)
	}

	return result, nil
}

type Order struct {
	Id                int64                  `bson:"id"`
	BizID             int64                  `bson:"biz_id"` // 业务ID
	Key               string                 `bson:"key"`    // 业务唯一 Key
	TemplateId        int64                  `bson:"template_id"`
	WorkflowId        int64                  `bson:"workflow_id"`
	ProcessInstanceId int                    `bson:"process_instance_id"`
	CreateBy          string                 `bson:"create_by"`
	Provide           uint8                  `bson:"provide"`
	Data              map[string]interface{} `bson:"data"`
	Status            uint8                  `bson:"status"`
	Ctime             int64                  `bson:"ctime"`
	Wtime             int64                  `bson:"wtime"`
	Utime             int64                  `bson:"utime"`
	NotificationConf  NotificationConf       `bson:"notification_conf"`
}

type NotificationConf struct {
	TemplateID     int64                  `bson:"template_id"`     // 模版ID
	TemplateParams map[string]interface{} `bson:"template_params"` // 传递参数
	Channel        string                 `bson:"channel"`         // 通知渠道
}
