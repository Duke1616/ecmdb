package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	TaskDataCollection = "c_order_task_forms"
)

type TaskFormDAO interface {
	// Create 创建任务快照
	Create(ctx context.Context, forms []TaskForm) error

	// FindByTaskIds 批量查询任务快照
	FindByTaskIds(ctx context.Context, taskIds []int) ([]TaskForm, error)

	// FindByOrderID 根据工单ID获取
	FindByOrderID(ctx context.Context, orderID int64) ([]TaskForm, error)
}

type taskFormDAO struct {
	db *mongox.Mongo
}

func NewTaskFormDAO(db *mongox.Mongo) TaskFormDAO {
	return &taskFormDAO{
		db: db,
	}
}

func (dao *taskFormDAO) Create(ctx context.Context, forms []TaskForm) error {
	if len(forms) == 0 {
		return nil
	}

	now := time.Now().UnixMilli()
	docs := make([]interface{}, 0, len(forms))
	for _, form := range forms {
		form.Id = dao.db.GetIdGenerator(TaskDataCollection)
		form.CreateTime = now
		docs = append(docs, form)
	}

	col := dao.db.Collection(TaskDataCollection)
	_, err := col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("批量插入任务表单数据失败: %w", err)
	}
	return nil
}

func (dao *taskFormDAO) FindByTaskIds(ctx context.Context, taskIds []int) ([]TaskForm, error) {
	col := dao.db.Collection(TaskDataCollection)
	filter := bson.M{"task_id": bson.M{"$in": taskIds}}

	cursor, err := col.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("查询任务表单数据失败: %w", err)
	}

	var res []TaskForm
	if err = cursor.All(ctx, &res); err != nil {
		return nil, fmt.Errorf("解码任务表单数据失败: %w", err)
	}
	return res, nil
}

func (dao *taskFormDAO) FindByOrderID(ctx context.Context, orderID int64) ([]TaskForm, error) {
	col := dao.db.Collection(TaskDataCollection)

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "order_id", Value: orderID}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "ctime", Value: -1}}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$key"},
			{Key: "result", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}},
		}}},
		bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$result"}}}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("查询任务表单数据失败: %w", err)
	}
	defer cursor.Close(ctx)

	var res []TaskForm
	if err = cursor.All(ctx, &res); err != nil {
		return nil, fmt.Errorf("解码任务表单数据失败: %w", err)
	}
	return res, nil
}

type TaskForm struct {
	Id         int64       `bson:"id"`
	OrderId    int64       `bson:"order_id"`
	TaskId     int         `bson:"task_id"`
	Name       string      `bson:"name"`
	Key        string      `bson:"key"`
	Type       string      `bson:"type"`
	Value      interface{} `bson:"value"`
	CreateTime int64       `bson:"ctime"`
}
