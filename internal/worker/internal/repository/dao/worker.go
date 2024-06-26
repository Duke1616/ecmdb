package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// ErrDataNotFound 通用的数据没找到
var ErrDataNotFound = mongo.ErrNoDocuments

const (
	WorkerCollection = "c_worker"
)

type WorkerDAO interface {
	CreateWorker(ctx context.Context, t Worker) (int64, error)
	FindByName(ctx context.Context, name string) (Worker, error)
	FindByKey(ctx context.Context, key string) (Worker, error)
	UpdateStatus(ctx context.Context, id int64, status uint8) (int64, error)
	ListWorker(ctx context.Context, offset, limit int64) ([]Worker, error)
	ListWorkerTopic(ctx context.Context) ([]Worker, error)
	Count(ctx context.Context) (int64, error)
}

func NewWorkerDAO(db *mongox.Mongo) WorkerDAO {
	return &workerDAO{
		db: db,
	}
}

type workerDAO struct {
	db *mongox.Mongo
}

func (dao *workerDAO) CreateWorker(ctx context.Context, t Worker) (int64, error) {
	t.Id = dao.db.GetIdGenerator(WorkerCollection)
	col := dao.db.Collection(WorkerCollection)
	now := time.Now()
	t.Ctime, t.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, t)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return t.Id, nil
}

func (dao *workerDAO) FindByName(ctx context.Context, name string) (Worker, error) {
	col := dao.db.Collection(WorkerCollection)
	var w Worker
	filter := bson.M{"name": name}

	if err := col.FindOne(ctx, filter).Decode(&w); err != nil {
		return Worker{}, fmt.Errorf("解码错误，%w", err)
	}

	return w, nil
}

func (dao *workerDAO) FindByKey(ctx context.Context, key string) (Worker, error) {
	col := dao.db.Collection(WorkerCollection)
	var w Worker
	filter := bson.M{"key": key}

	if err := col.FindOne(ctx, filter).Decode(&w); err != nil {
		return Worker{}, fmt.Errorf("解码错误，%w", err)
	}

	return w, nil
}

func (dao *workerDAO) UpdateStatus(ctx context.Context, id int64, status uint8) (int64, error) {
	col := dao.db.Collection(WorkerCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"status": status,
			"utime":  time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *workerDAO) ListWorker(ctx context.Context, offset, limit int64) ([]Worker, error) {
	col := dao.db.Collection(WorkerCollection)
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

	var result []Worker
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *workerDAO) Count(ctx context.Context) (int64, error) {
	col := dao.db.Collection(WorkerCollection)
	filer := bson.M{}

	count, err := col.CountDocuments(ctx, filer)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *workerDAO) ListWorkerTopic(ctx context.Context) ([]Worker, error) {
	col := dao.db.Collection(WorkerCollection)
	filter := bson.M{}
	projection := bson.M{"topic": 1}
	cursor, err := col.Find(ctx, filter, options.Find().SetProjection(projection))
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []Worker
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

type Worker struct {
	Id     int64  `json:"id"`
	Key    string `json:"key"`
	Name   string `bson:"name"`
	Topic  string `bson:"topic"`
	Desc   string `bson:"desc"`
	Status uint8  `bson:"status"`
	Ctime  int64  `bson:"ctime"`
	Utime  int64  `bson:"utime"`
}
