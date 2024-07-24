package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const TaskCollection = "c_task"

type TaskDAO interface {
	CreateTask(ctx context.Context, t Task) (int64, error)
	UpdateTaskStatus(ctx context.Context, req Task) (int64, error)
	ListTask(ctx context.Context, offset, limit int64) ([]Task, error)
	Count(ctx context.Context) (int64, error)
}

type taskDAO struct {
	db *mongox.Mongo
}

func NewTaskDAO(db *mongox.Mongo) TaskDAO {
	return &taskDAO{
		db: db,
	}
}

func (dao *taskDAO) UpdateTaskStatus(ctx context.Context, t Task) (int64, error) {
	col := dao.db.Collection(TaskCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"result": t.Result,
			"status": t.Status,
			"utime":  time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": t.Id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *taskDAO) CreateTask(ctx context.Context, t Task) (int64, error) {
	now := time.Now()
	t.Ctime, t.Utime = now.UnixMilli(), now.UnixMilli()
	t.Id = dao.db.GetIdGenerator(TaskCollection)
	col := dao.db.Collection(TaskCollection)

	_, err := col.InsertOne(ctx, t)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return t.Id, nil
}

func (dao *taskDAO) ListTask(ctx context.Context, offset, limit int64) ([]Task, error) {
	col := dao.db.Collection(TaskCollection)
	filter := bson.M{}
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

	var result []Task
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *taskDAO) Count(ctx context.Context) (int64, error) {
	col := dao.db.Collection(TaskCollection)
	filer := bson.M{}

	count, err := col.CountDocuments(ctx, filer)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

type Task struct {
	Id            int64  `bson:"id"`
	OrderId       int64  `json:"order_id"`
	ProcessInstId int    `bson:"process_inst_id"`
	CodebookUid   string `bson:"codebook_uid"`
	WorkerName    string `bson:"worker_name"`
	WorkflowId    int64  `bson:"workflow_id"`
	Code          string `bson:"code"`
	Topic         string `bson:"topic"`
	Language      string `bson:"language"`
	Status        uint8  `bson:"status"`
	Result        string `bson:"result"`
	Ctime         int64  `bson:"ctime"`
	Utime         int64  `bson:"utime"`
}
