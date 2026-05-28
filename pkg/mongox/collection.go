package mongox

import (
	"context"
	"fmt"
	"strings"

	"github.com/Duke1616/eiam/pkg/ctxutil"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection 泛型强类型数据集包装器
type Collection[T any] struct {
	db   *DB
	coll *mongo.Collection
	name string
}

// NewCollection 实例化泛型 Collection
func NewCollection[T any](db *DB, collName string) *Collection[T] {
	return &Collection[T]{
		db:   db,
		coll: db.native.Collection(collName),
		name: collName,
	}
}

// Native 直接获取底层的 *mongo.Collection 以进行操作
func (c *Collection[T]) Native() *mongo.Collection {
	return c.coll
}

// ==========================================
// 核心泛型 CRUD API
// ==========================================

// FindOne 查询单条记录，应用生命周期拦截并自动反序列化
func (c *Collection[T]) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*T, error) {
	if c.shouldShortCircuit(ctx) {
		return nil, mongo.ErrNoDocuments
	}
	var dest T
	finalFilter, err := c.applyBeforeFind(ctx, filter, &dest)
	if err != nil {
		return nil, err
	}

	err = c.coll.FindOne(ctx, finalFilter, opts...).Decode(&dest)
	if err != nil {
		return nil, err
	}
	return &dest, nil
}

// Find 查询多条记录，应用生命周期拦截并自动转换为强类型切片
func (c *Collection[T]) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]T, error) {
	if c.shouldShortCircuit(ctx) {
		return []T{}, nil
	}
	var dest []T
	finalFilter, err := c.applyBeforeFind(ctx, filter, &dest)
	if err != nil {
		return nil, err
	}

	cursor, err := c.coll.Find(ctx, finalFilter, opts...)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	results := make([]T, 0)
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// InsertOne 写入单条文档，执行 BeforeInsert 拦截（自动填充 TenantID 与自增 ID）
func (c *Collection[T]) InsertOne(ctx context.Context, doc *T, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	stmt := &Statement{
		CollectionName: c.name,
		Model:          doc,
		Context:        ctx,
	}
	if err := c.db.runBeforeInsert(stmt); err != nil {
		return nil, err
	}
	return c.coll.InsertOne(ctx, doc, opts...)
}

// InsertMany 批量写入文档，自动强转并运行批量 ID 审计拦截
func (c *Collection[T]) InsertMany(ctx context.Context, docs []*T, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	stmt := &Statement{
		CollectionName: c.name,
		Model:          &docs,
		Context:        ctx,
	}
	if err := c.db.runBeforeInsert(stmt); err != nil {
		return nil, err
	}

	interfaceDocs := make([]interface{}, len(docs))
	for i, d := range docs {
		interfaceDocs[i] = d
	}
	return c.coll.InsertMany(ctx, interfaceDocs, opts...)
}

// UpdateOne 更新单条文档，自动拦截并追加防越权租户隔离条件
func (c *Collection[T]) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	stmt := &Statement{
		CollectionName: c.name,
		Model:          new(T),
		Filter:         toFilterMap(filter),
		Update:         update,
		Context:        ctx,
	}
	if err := c.db.runBeforeUpdate(stmt); err != nil {
		return nil, err
	}
	return c.coll.UpdateOne(ctx, stmt.Filter, stmt.Update, opts...)
}

// UpdateMany 批量更新文档，自动拦截并追加租户隔离条件
func (c *Collection[T]) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	stmt := &Statement{
		CollectionName: c.name,
		Model:          new(T),
		Filter:         toFilterMap(filter),
		Update:         update,
		Context:        ctx,
	}
	if err := c.db.runBeforeUpdate(stmt); err != nil {
		return nil, err
	}
	return c.coll.UpdateMany(ctx, stmt.Filter, stmt.Update, opts...)
}

// DeleteOne 删除单条文档，自动拦截并追加写隔离校验
func (c *Collection[T]) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	stmt := &Statement{
		CollectionName: c.name,
		Model:          new(T),
		Filter:         toFilterMap(filter),
		Context:        ctx,
	}
	if err := c.db.runBeforeDelete(stmt); err != nil {
		return nil, err
	}
	return c.coll.DeleteOne(ctx, stmt.Filter, opts...)
}

// DeleteMany 批量删除文档，自动拦截并追加写隔离校验
func (c *Collection[T]) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	stmt := &Statement{
		CollectionName: c.name,
		Model:          new(T),
		Filter:         toFilterMap(filter),
		Context:        ctx,
	}
	if err := c.db.runBeforeDelete(stmt); err != nil {
		return nil, err
	}
	return c.coll.DeleteMany(ctx, stmt.Filter, opts...)
}

// CountDocuments 统计文档数量，自动触发 BeforeFinder 拦截
func (c *Collection[T]) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	if c.shouldShortCircuit(ctx) {
		return 0, nil
	}
	finalFilter, err := c.applyBeforeFind(ctx, filter, nil)
	if err != nil {
		return 0, err
	}
	return c.coll.CountDocuments(ctx, finalFilter, opts...)
}

// Distinct 获取去重字段集，自动拦截查询并追加租户隔离
func (c *Collection[T]) Distinct(ctx context.Context, fieldName string, filter interface{}, opts ...*options.DistinctOptions) ([]interface{}, error) {
	if c.shouldShortCircuit(ctx) {
		return []interface{}{}, nil
	}
	finalFilter, err := c.applyBeforeFind(ctx, filter, nil)
	if err != nil {
		return nil, err
	}
	return c.coll.Distinct(ctx, fieldName, finalFilter, opts...)
}

// Aggregate 聚合查询方法，自动在管道最前端织入租户过滤阶段以保障租户隔离
func (c *Collection[T]) Aggregate(ctx context.Context, pipeline mongo.Pipeline, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	tenantID := ctxutil.GetTenantID(ctx).Int64()
	if tenantID < 0 {
		tenantID = 0
	}
	ignoreTenant, _ := ctx.Value("mongox:ignore_tenant").(bool)

	// 只要没有显式开启受控豁免，在管道最前端织入当前租户空间的限制过滤条件
	if !ignoreTenant {
		tenantMatch := bson.D{{
			Key: "$match",
			Value: bson.M{
				"tenant_id": tenantID,
			},
		}}
		pipeline = append(mongo.Pipeline{tenantMatch}, pipeline...)
	}

	return c.coll.Aggregate(ctx, pipeline, opts...)
}

// BulkWrite 批量写入/更新/删除操作，自动为每个子操作应用多租户插件
func (c *Collection[T]) BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	if len(models) == 0 {
		return &mongo.BulkWriteResult{}, nil
	}

	for _, m := range models {
		switch model := m.(type) {
		case *mongo.InsertOneModel:
			stmt := &Statement{
				CollectionName: c.name,
				Model:          model.Document,
				Context:        ctx,
			}
			if err := c.db.runBeforeInsert(stmt); err != nil {
				return nil, err
			}
			model.Document = stmt.Model

		case *mongo.UpdateOneModel:
			stmt := &Statement{
				CollectionName: c.name,
				Model:          new(T),
				Filter:         toFilterMap(model.Filter),
				Update:         model.Update,
				Context:        ctx,
			}
			if err := c.db.runBeforeUpdate(stmt); err != nil {
				return nil, err
			}
			model.SetFilter(stmt.Filter)

		case *mongo.UpdateManyModel:
			stmt := &Statement{
				CollectionName: c.name,
				Model:          new(T),
				Filter:         toFilterMap(model.Filter),
				Update:         model.Update,
				Context:        ctx,
			}
			if err := c.db.runBeforeUpdate(stmt); err != nil {
				return nil, err
			}
			model.SetFilter(stmt.Filter)

		case *mongo.DeleteOneModel:
			stmt := &Statement{
				CollectionName: c.name,
				Model:          new(T),
				Filter:         toFilterMap(model.Filter),
				Context:        ctx,
			}
			if err := c.db.runBeforeDelete(stmt); err != nil {
				return nil, err
			}
			model.SetFilter(stmt.Filter)

		case *mongo.DeleteManyModel:
			stmt := &Statement{
				CollectionName: c.name,
				Model:          new(T),
				Filter:         toFilterMap(model.Filter),
				Context:        ctx,
			}
			if err := c.db.runBeforeDelete(stmt); err != nil {
				return nil, err
			}
			model.SetFilter(stmt.Filter)

		case *mongo.ReplaceOneModel:
			stmtUpdate := &Statement{
				CollectionName: c.name,
				Model:          new(T),
				Filter:         toFilterMap(model.Filter),
				Context:        ctx,
			}
			if err := c.db.runBeforeUpdate(stmtUpdate); err != nil {
				return nil, err
			}
			model.SetFilter(stmtUpdate.Filter)

			stmtInsert := &Statement{
				CollectionName: c.name,
				Model:          model.Replacement,
				Context:        ctx,
			}
			if err := c.db.runBeforeInsert(stmtInsert); err != nil {
				return nil, err
			}
			model.Replacement = stmtInsert.Model
		}
	}

	return c.coll.BulkWrite(ctx, models, opts...)
}

// ==========================================
// 插件生命周期辅助方法
// ==========================================

func (c *Collection[T]) applyBeforeFind(ctx context.Context, filter interface{}, dest interface{}) (bson.M, error) {
	model := dest
	if model == nil {
		model = new(T)
	}
	stmt := &Statement{
		CollectionName: c.name,
		Model:          model,
		Filter:         toFilterMap(filter),
		Context:        ctx,
	}
	if err := c.db.runBeforeFind(stmt); err != nil {
		return nil, err
	}
	return stmt.Filter, nil
}

func (c *Collection[T]) shouldShortCircuit(ctx context.Context) bool {
	tid := ctxutil.GetTenantID(ctx).Int64()
	ignoreTenant, _ := ctx.Value("mongox:ignore_tenant").(bool)
	return tid <= 0 && !ignoreTenant
}

func toFilterMap(filter interface{}) bson.M {
	if filter == nil {
		return bson.M{}
	}
	if m, ok := filter.(bson.M); ok {
		return m
	}
	byteData, err := bson.Marshal(filter)
	if err != nil {
		return bson.M{}
	}
	var result bson.M
	if err = bson.Unmarshal(byteData, &result); err != nil {
		return bson.M{}
	}
	return result
}

// SyncIndexes 声明式索引同步器，对比活跃索引与代码预期配置，自动增删对齐
func SyncIndexes(ctx context.Context, col *mongo.Collection, expected []mongo.IndexModel) error {
	cursor, err := col.Indexes().List(ctx)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	type activeIndex struct {
		Name   string `bson:"name"`
		Key    bson.M `bson:"key"`
		Unique bool   `bson:"unique"`
	}

	var active []activeIndex
	if err = cursor.All(ctx, &active); err != nil {
		return err
	}

	activeMap := lo.Associate(
		lo.Filter(active, func(idx activeIndex, _ int) bool {
			return idx.Name != "_id_"
		}),
		func(idx activeIndex) (string, activeIndex) {
			return idx.Name, idx
		},
	)

	expectedMap := lo.Associate(expected, func(exp mongo.IndexModel) (string, mongo.IndexModel) {
		var name string
		if exp.Options != nil && exp.Options.Name != nil {
			name = *exp.Options.Name
		} else {
			name = generateIndexName(exp.Keys)
		}
		return name, exp
	})

	for currentName, currentIdx := range activeMap {
		exp, exists := expectedMap[currentName]
		if !exists {
			_, _ = col.Indexes().DropOne(ctx, currentName)
			continue
		}

		expUnique := false
		if exp.Options != nil && exp.Options.Unique != nil {
			expUnique = *exp.Options.Unique
		}
		if currentIdx.Unique != expUnique {
			_, _ = col.Indexes().DropOne(ctx, currentName)
		}
	}

	if len(expected) > 0 {
		_, err = col.Indexes().CreateMany(ctx, expected)
		if err != nil {
			return err
		}
	}

	return nil
}

// generateIndexName 根据 bson.D 键定义生成默认的索引名
func generateIndexName(keys interface{}) string {
	if keys == nil {
		return ""
	}
	var d bson.D
	switch k := keys.(type) {
	case bson.D:
		d = k
	default:
		bytes, err := bson.Marshal(keys)
		if err == nil {
			_ = bson.Unmarshal(bytes, &d)
		}
	}

	parts := lo.Map(d, func(elem bson.E, _ int) string {
		return fmt.Sprintf("%s_%v", elem.Key, elem.Value)
	})
	return strings.Join(parts, "_")
}
