package dao

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type ModelDAO interface {
	CreateModelGroup(ctx context.Context, mg ModelGroup) (int64, error)
}

func NewModelDAO(client *mongo.Client) ModelDAO {
	return &modelDAO{
		db: client,
	}
}

type modelDAO struct {
	db *mongo.Client
}

func (m *modelDAO) CreateModelGroup(ctx context.Context, mg ModelGroup) (int64, error) {
	now := time.Now()
	mg.Ctime, mg.Utime = now.UnixMilli(), now.UnixMilli()

	col := m.db.Database("cmdb").
		Collection("t_model_group")

	result, err := col.InsertOne(ctx, mg)

	if err != nil {
		return 0, err
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return 0, errors.New("inserted ID is not of type primitive.ObjectID")
	}

	// 将 ObjectID 转换为 Unix 时间戳
	return id.Timestamp().Unix(), nil
}

type ModelGroup struct {
	Id    int64  `bson:"_id"`
	Name  string `bson:"name"`
	Ctime int64  `bson:"ctime"`
	Utime int64  `bson:"utime"`
}
