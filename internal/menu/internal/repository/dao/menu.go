package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const MenuCollection = "c_menu"

type MenuDAO interface {
	CreateMenu(ctx context.Context, t Menu) (int64, error)
	UpdateMenu(ctx context.Context, t Menu) (int64, error)
	ListMenu(ctx context.Context) ([]Menu, error)
}

type menuDAO struct {
	db *mongox.Mongo
}

func (dao *menuDAO) ListMenu(ctx context.Context) ([]Menu, error) {
	col := dao.db.Collection(MenuCollection)
	filter := bson.M{}

	opts := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}
	cursor, err := col.Find(ctx, filter, opts)
	var result []Menu
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *menuDAO) UpdateMenu(ctx context.Context, t Menu) (int64, error) {
	col := dao.db.Collection(MenuCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"name":           t.Name,
			"pid":            t.Pid,
			"path":           t.Path,
			"sort":           t.Sort,
			"redirect":       t.Redirect,
			"component":      t.Component,
			"component_path": t.ComponentPath,
			"type":           t.Type,
			"status":         t.Status,
			"meta":           t.Meta,
			"endpoints":      t.Endpoints,
			"utime":          time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": t.Id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *menuDAO) CreateMenu(ctx context.Context, e Menu) (int64, error) {
	e.Id = dao.db.GetIdGenerator(MenuCollection)
	col := dao.db.Collection(MenuCollection)
	now := time.Now()
	e.Ctime, e.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, e)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return e.Id, nil
}

func NewMenuDAO(db *mongox.Mongo) MenuDAO {
	return &menuDAO{
		db: db,
	}
}

type Menu struct {
	Id            int64      `bson:"id"`
	Pid           int64      `bson:"pid"`
	Path          string     `bson:"path"`
	Sort          int64      `bson:"sort"`
	Name          string     `bson:"name"`
	Redirect      string     `bson:"redirect"`
	Component     string     `bson:"component"`
	ComponentPath string     `bson:"component_path"`
	Type          uint8      `bson:"type"`
	Status        uint8      `bson:"status"`
	Meta          Meta       `bson:"meta"`
	Endpoints     []Endpoint `bson:"endpoints"`
	Ctime         int64      `bson:"ctime"`
	Utime         int64      `bson:"utime"`
}

type Endpoint struct {
	Id     int64  `bson:"id"`
	Path   string `bson:"path"`
	Method string `bson:"method"`
	Desc   string `bson:"desc"`
}

type Meta struct {
	Title       string `bson:"title"`        // 展示名称
	IsHidden    bool   `bson:"is_hidden"`    // 是否展示
	IsAffix     bool   `bson:"is_affix"`     // 是否固定
	IsKeepAlive bool   `bson:"is_keepalive"` // 是否缓存
	Icon        string `bson:"icon"`         // Icon图标
}
