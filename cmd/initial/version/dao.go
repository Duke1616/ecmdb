package version

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Dao interface {
	CreateOrUpdateVersion(ctx context.Context, version string) error
	GetVersion(ctx context.Context) (string, error)
}

type dao struct {
	db *mongox.Mongo
}

func (d *dao) GetVersion(ctx context.Context) (string, error) {
	col := d.db.Collection("c_version")
	var result Version
	filter := bson.M{}

	if err := col.FindOne(ctx, filter).Decode(&result); err != nil {
		return result.CurrentVersion, fmt.Errorf("解码错误，%w", err)
	}

	return "", nil
}

func NewDao(db *mongox.Mongo) Dao {
	return &dao{
		db: db,
	}
}

func (d *dao) CreateOrUpdateVersion(ctx context.Context, version string) error {
	col := d.db.Collection("c_version")
	filter := bson.M{}
	updateDoc := bson.M{
		"$set": bson.M{
			"current_version": version,
			"utime":           time.Now().UnixMilli(),
		},
	}

	upsert := true
	opts := &options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := col.UpdateOne(ctx, filter, updateDoc, opts)
	return err
}

type Version struct {
	// 当前服务初始化数据版本
	CurrentVersion string `json:"current_version" bson:"current_version"`
	Ctime          int64  `bson:"ctime"`
	Utime          int64  `bson:"utime"`
}
