package version

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Dao interface {
	CreateOrUpdateVersion(ctx context.Context, version string) error
	GetVersion(ctx context.Context) (string, error)
	// 菜单相关方法
	SetMenuHash(ctx context.Context, hash string) error
	GetMenuHash(ctx context.Context) (string, error)
}

type dao struct {
	db *mongox.Mongo
}

func (d *dao) GetVersion(ctx context.Context) (string, error) {
	col := d.db.Collection("c_version")
	var result Version
	filter := bson.M{}

	if err := col.FindOne(ctx, filter).Decode(&result); err != nil {
		return "", fmt.Errorf("解码错误，%w", err)
	}

	return result.CurrentVersion, nil
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

func (d *dao) SetMenuHash(ctx context.Context, hash string) error {
	col := d.db.Collection("c_version")
	filter := bson.M{} // 更新当前版本记录
	updateDoc := bson.M{
		"$set": bson.M{
			"menu_hash": hash,
			"utime":     time.Now().UnixMilli(),
		},
	}

	upsert := true
	opts := &options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := col.UpdateOne(ctx, filter, updateDoc, opts)
	return err
}

func (d *dao) GetMenuHash(ctx context.Context) (string, error) {
	col := d.db.Collection("c_version")
	var result Version
	filter := bson.M{} // 获取当前版本记录

	if err := col.FindOne(ctx, filter).Decode(&result); err != nil {
		return "", fmt.Errorf("获取菜单哈希失败，%w", err)
	}

	return result.MenuHash, nil
}

type Version struct {
	// 当前服务初始化数据版本
	CurrentVersion string `json:"current_version" bson:"current_version"`
	// 菜单文件 MD5 哈希值
	MenuHash string `json:"menu_hash" bson:"menu_hash"`
	Ctime    int64  `bson:"ctime"`
	Utime    int64  `bson:"utime"`
}
