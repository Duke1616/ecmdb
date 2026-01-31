package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	TaskDataCollection = "order_snapshots"
)

type OrderSnapshotsDAO interface {
	Create(ctx context.Context, data TaskData) error
	FindByTaskIds(ctx context.Context, taskIds []int) ([]TaskData, error)
}

type orderSnapshotsDAO struct {
	db *mongox.Mongo
}

func NewOrderSnapshotsDAO(db *mongox.Mongo) OrderSnapshotsDAO {
	return &orderSnapshotsDAO{
		db: db,
	}
}

func (dao *orderSnapshotsDAO) Create(ctx context.Context, data TaskData) error {
	data.CreateTime = time.Now().UnixMilli()
	data.Id = dao.db.GetIdGenerator(TaskDataCollection)
	col := dao.db.Collection(TaskDataCollection)

	_, err := col.InsertOne(ctx, data)
	if err != nil {
		return fmt.Errorf("插入任务数据快照失败: %w", err)
	}
	return nil
}

func (dao *orderSnapshotsDAO) FindByTaskIds(ctx context.Context, taskIds []int) ([]TaskData, error) {
	col := dao.db.Collection(TaskDataCollection)
	filter := bson.M{"task_id": bson.M{"$in": taskIds}}

	cursor, err := col.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("查询任务数据失败: %w", err)
	}

	var res []TaskData
	if err = cursor.All(ctx, &res); err != nil {
		return nil, fmt.Errorf("解码任务数据失败: %w", err)
	}
	return res, nil
}

type TaskData struct {
	Id         int64                  `bson:"id"`
	TaskId     int                    `bson:"task_id"`
	Data       map[string]interface{} `bson:"data"`
	CreateTime int64                  `bson:"ctime"`
}
