package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

const TaskCollection = "c_task"

type TaskDAO interface {
	CreateTask(ctx context.Context, t Task) (int64, error)
	UpdateTaskStatus(ctx context.Context, req Task) (int64, error)
}

type taskDAO struct {
	db *mongox.Mongo
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

func NewTaskDAO(db *mongox.Mongo) TaskDAO {
	return &taskDAO{
		db: db,
	}
}

type Task struct {
	Id            int64  `bson:"id"`
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
