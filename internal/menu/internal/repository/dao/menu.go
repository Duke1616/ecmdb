package dao

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/ekit/slice"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const MenuCollection = "c_menu"

type MenuDAO interface {
	CreateMenu(ctx context.Context, t Menu) (int64, error)
	UpdateMenu(ctx context.Context, t Menu) (int64, error)
	ListMenu(ctx context.Context) ([]Menu, error)
	// ListByPlatform 根据平台获取菜单列表
	ListByPlatform(ctx context.Context, platform string) ([]Menu, error)
	FindByIds(ctx context.Context, ids []int64) ([]Menu, error)
	GetAllMenu(ctx context.Context) ([]Menu, error)
	FindById(ctx context.Context, id int64) (Menu, error)
	DeleteMenu(ctx context.Context, id int64) (int64, error)

	// InjectMenu 注入菜单数据
	InjectMenu(ctx context.Context, ms []Menu) (*mongo.BulkWriteResult, error)

	// UpdateMenuEndpoints 同步菜单API数据变更
	UpdateMenuEndpoints(ctx context.Context, id int64, endpoints []Endpoint) (int64, error)
}

type menuDAO struct {
	db *mongox.Mongo
}

func (dao *menuDAO) UpdateMenuEndpoints(ctx context.Context, id int64, endpoints []Endpoint) (int64, error) {
	col := dao.db.Collection(MenuCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"endpoints": endpoints,
			"utime":     time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *menuDAO) ListByPlatform(ctx context.Context, platform string) ([]Menu, error) {
	col := dao.db.Collection(MenuCollection)
	filter := bson.M{}
	if platform != "" {
		filter["meta.platforms"] = bson.M{
			"$elemMatch": bson.M{"$eq": platform},
		}
	}
	cursor, err := col.Find(ctx, filter)
	var result []Menu
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *menuDAO) DeleteMenu(ctx context.Context, id int64) (int64, error) {
	col := dao.db.Collection(MenuCollection)
	filter := bson.M{"id": id}

	result, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *menuDAO) FindById(ctx context.Context, id int64) (Menu, error) {
	col := dao.db.Collection(MenuCollection)
	filter := bson.M{"id": id}

	var result Menu
	if err := col.FindOne(ctx, filter).Decode(&result); err != nil {
		return Menu{}, fmt.Errorf("解码错误，%w", err)
	}

	return result, nil
}

func (dao *menuDAO) GetAllMenu(ctx context.Context) ([]Menu, error) {
	col := dao.db.Collection(MenuCollection)
	filter := bson.M{}

	cursor, err := col.Find(ctx, filter)
	var result []Menu
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *menuDAO) FindByIds(ctx context.Context, ids []int64) ([]Menu, error) {
	col := dao.db.Collection(MenuCollection)
	filter := bson.M{"id": bson.M{"$in": ids}}

	cursor, err := col.Find(ctx, filter)
	var result []Menu
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
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

func (dao *menuDAO) InjectMenu(ctx context.Context, ms []Menu) (*mongo.BulkWriteResult, error) {
	col := dao.db.Collection(MenuCollection)
	now := time.Now()

	// 先把库里已有的 menu 取出来，映射成 map
	ids := slice.Map(ms, func(idx int, src Menu) int64 {
		return src.Id
	})
	cur, err := col.Find(ctx, bson.M{"id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	existing := make(map[int64]Menu)
	for cur.Next(ctx) {
		var m Menu
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		existing[m.Id] = m
	}

	// 构造 bulk operations
	operations := make([]mongo.WriteModel, 0, len(ms))
	for _, menu := range ms {
		menu.Ctime = now.UnixMilli()
		menu.Utime = now.UnixMilli()

		old, ok := existing[menu.Id]

		// 计算差异
		setFields := bson.M{}
		if !ok || old.Pid != menu.Pid {
			setFields["pid"] = menu.Pid
		}
		if !ok || old.Path != menu.Path {
			setFields["path"] = menu.Path
		}
		if !ok || old.Name != menu.Name {
			setFields["name"] = menu.Name
		}
		if !ok || old.Sort != menu.Sort {
			setFields["sort"] = menu.Sort
		}
		if !ok || old.Component != menu.Component {
			setFields["component"] = menu.Component
		}
		if !ok || old.Redirect != menu.Redirect {
			setFields["redirect"] = menu.Redirect
		}
		if !ok || old.Status != menu.Status {
			setFields["status"] = menu.Status
		}
		if !ok || old.Type != menu.Type {
			setFields["type"] = menu.Type
		}
		if !ok || !reflect.DeepEqual(old.Meta, menu.Meta) {
			setFields["meta"] = menu.Meta
		}
		if !ok || !reflect.DeepEqual(old.Endpoints, menu.Endpoints) {
			setFields["endpoints"] = menu.Endpoints
		}

		// 如果字段有变化，才更新 utime
		if len(setFields) > 0 {
			setFields["utime"] = menu.Utime
		}

		// 构造 update 文档
		updateDoc := bson.M{}
		if len(setFields) > 0 {
			updateDoc["$set"] = setFields
		}
		updateDoc["$setOnInsert"] = bson.M{"ctime": menu.Ctime}

		operations = append(operations, &mongo.UpdateOneModel{
			Filter: bson.M{"id": menu.Id},
			Update: updateDoc,
			Upsert: &[]bool{true}[0],
		})
	}

	if len(operations) == 0 {
		// 没有任何需要写入的变化
		return &mongo.BulkWriteResult{}, nil
	}

	return col.BulkWrite(ctx, operations)
}

func (dao *menuDAO) UpdateMenu(ctx context.Context, t Menu) (int64, error) {
	col := dao.db.Collection(MenuCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"name":      t.Name,
			"pid":       t.Pid,
			"path":      t.Path,
			"sort":      t.Sort,
			"redirect":  t.Redirect,
			"component": t.Component,
			"type":      t.Type,
			"status":    t.Status,
			"meta":      t.Meta,
			"utime":     time.Now().UnixMilli(),
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
	Id        int64      `bson:"id"`
	Pid       int64      `bson:"pid"`
	Path      string     `bson:"path"`
	Sort      int64      `bson:"sort"`
	Name      string     `bson:"name"`
	Redirect  string     `bson:"redirect"`
	Component string     `bson:"component"`
	Type      uint8      `bson:"type"`
	Status    uint8      `bson:"status"`
	Meta      Meta       `bson:"meta"`
	Endpoints []Endpoint `bson:"endpoints"`
	Ctime     int64      `bson:"ctime"`
	Utime     int64      `bson:"utime"`
}

type Endpoint struct {
	Path     string `bson:"path"`
	Method   string `bson:"method"`
	Resource string `bson:"resource"`
	Desc     string `bson:"desc"`
}

type Meta struct {
	Title       string   `bson:"title"`        // 展示名称
	IsHidden    bool     `bson:"is_hidden"`    // 是否展示
	IsAffix     bool     `bson:"is_affix"`     // 是否固定
	Platforms   []string `bson:"platforms"`    // 作用平台
	IsKeepAlive bool     `bson:"is_keepalive"` // 是否缓存
	Icon        string   `bson:"icon"`         // Icon图标
}
