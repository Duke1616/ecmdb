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
	RunnerCollection = "c_runner"
)

type RunnerDAO interface {
	CreateRunner(ctx context.Context, r Runner) (int64, error)
	ListRunner(ctx context.Context, offset, limit int64) ([]Runner, error)
	Count(ctx context.Context) (int64, error)
}

func NewRunnerDAO(db *mongox.Mongo) RunnerDAO {
	return &runnerDAO{
		db: db,
	}
}

type runnerDAO struct {
	db *mongox.Mongo
}

func (dao *runnerDAO) CreateRunner(ctx context.Context, r Runner) (int64, error) {
	r.Id = dao.db.GetIdGenerator(RunnerCollection)
	col := dao.db.Collection(RunnerCollection)
	now := time.Now()
	r.Ctime, r.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, r)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return r.Id, nil
}

func (dao *runnerDAO) ListRunner(ctx context.Context, offset, limit int64) ([]Runner, error) {
	col := dao.db.Collection(RunnerCollection)
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

	var result []Runner
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *runnerDAO) Count(ctx context.Context) (int64, error) {
	col := dao.db.Collection(RunnerCollection)
	filer := bson.M{}

	count, err := col.CountDocuments(ctx, filer)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

type Runner struct {
	Id             int64    `bson:"id"`
	Name           string   `bson:"name"`
	TaskIdentifier string   `bson:"task_identifier"`
	TaskSecret     string   `bson:"task_secret"`
	WorkName       string   `bson:"work_name"`
	Tags           []string `bson:"tags"`
	Action         uint8    `bson:"action"`
	Desc           string   `json:"desc"`
	Ctime          int64    `bson:"ctime"`
	Utime          int64    `bson:"utime"`
}
