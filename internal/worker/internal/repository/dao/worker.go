package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/mongo"
)

// ErrDataNotFound 通用的数据没找到
var ErrDataNotFound = mongo.ErrNoDocuments

const (
	WorkerCollection = "c_worker"
)

type WorkerDAO interface {
	CreateWorker(ctx context.Context, t Worker) (int64, error)
	FindByName(ctx context.Context, name string) (Worker, error)
}

func NewWorkerDAO(db *mongox.Mongo) WorkerDAO {
	return &workerDAO{
		db: db,
	}
}

type workerDAO struct {
	db *mongox.Mongo
}

func (w workerDAO) CreateWorker(ctx context.Context, t Worker) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (w workerDAO) FindByName(ctx context.Context, name string) (Worker, error) {
	//TODO implement me
	panic("implement me")
}

type Worker struct {
	Name  string `bson:"name"`
	Topic string `bson:"topic"`
	Desc  string `bson:"desc"`
}
