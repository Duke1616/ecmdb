package plugin

import (
	"context"
	"errors"
	"reflect"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const idGeneratorCollection = "c_id_generator"

// idGeneratorRecord c_id_generator 集合中的文档结构
type idGeneratorRecord struct {
	Name   string `bson:"name"`
	NextID int64  `bson:"next_id"`
}

// AutoIDPlugin 自动为插入的文档分配全局唯一、单调递增的 ID。
//
// 单条插入：原子获取 1 个 ID 后立即注入文档。
// 批量插入：一次性原子申请 N 个连续 ID 后批量分发，消除 N 次串行网络往返。
type AutoIDPlugin struct {
	db *mongo.Database
}

// NewAutoIDPlugin 实例化自增 ID 插件
func NewAutoIDPlugin(db *mongo.Database) *AutoIDPlugin {
	return &AutoIDPlugin{db: db}
}

// Name 实现 Plugin 接口
func (p *AutoIDPlugin) Name() string {
	return "auto_id_plugin"
}

// nextID 从 c_id_generator 中原子地为指定集合申请 count 个连续自增 ID，返回起始 ID
func (p *AutoIDPlugin) nextID(ctx context.Context, collName string, count int64) (int64, error) {
	if p.db == nil {
		return 0, errors.New("auto_id_plugin: 数据库实例未初始化")
	}
	coll := p.db.Collection(idGeneratorCollection)

	upsert := true
	returnAfter := options.After
	result := idGeneratorRecord{}

	err := coll.FindOneAndUpdate(
		ctx,
		bson.M{"name": collName},
		bson.M{"$inc": bson.M{"next_id": count}},
		&options.FindOneAndUpdateOptions{
			Upsert:         &upsert,
			ReturnDocument: &returnAfter,
		},
	).Decode(&result)

	if err != nil {
		return 0, err
	}

	// 起始 ID = NextID - count + 1
	return result.NextID - count + 1, nil
}

// BeforeInsert 实现 BeforeInserter 接口，拦截写入，自动识别单条或批量切片并分配 ID
func (p *AutoIDPlugin) BeforeInsert(stmt *mongox.Statement) error {
	if stmt.Model == nil {
		return nil
	}

	v := reflect.ValueOf(stmt.Model)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		// 批量写入场景
		var needIDCount int64
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if model, ok := item.Interface().(mongox.IModel); ok && model.GetID() == 0 {
				needIDCount++
			}
		}
		if needIDCount == 0 {
			return nil
		}

		startID, err := p.nextID(stmt.Context, stmt.CollectionName, needIDCount)
		if err != nil {
			return err
		}

		var cursor int64
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if model, ok := item.Interface().(mongox.IModel); ok && model.GetID() == 0 {
				model.SetID(startID + cursor)
				cursor++
			}
		}

	default:
		// 单条写入场景
		if model, ok := stmt.Model.(mongox.IModel); ok && model.GetID() == 0 {
			id, err := p.nextID(stmt.Context, stmt.CollectionName, 1)
			if err != nil {
				return err
			}
			model.SetID(id)
		}
	}

	return nil
}
