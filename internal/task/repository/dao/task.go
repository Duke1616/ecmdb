package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
)

const TaskCollection = "c_task"

type TaskDAO interface {
	CreateTask(ctx context.Context, resource Task) (int64, error)
}

type taskDAO struct {
	db *mongox.Mongo
}

func NewTaskDAO(db *mongox.Mongo) TaskDAO {
	return &taskDAO{
		db: db,
	}
}

func (t *taskDAO) CreateTask(ctx context.Context, resource Task) (int64, error) {
	//TODO implement me
	panic("implement me")
}

type Task struct {
	Id       int64  `bson:"id"`
	ModelUID string `bson:"model_uid"`
	Data     string `bson:"data"`
	Ctime    int64  `bson:"ctime"`
	Utime    int64  `bson:"utime"`
}
