package mongox

import (
	"context"

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

// NewCollection 实例化强类型泛型 Collection
func NewCollection[T any](db *DB, collName string) *Collection[T] {
	return &Collection[T]{
		db:   db,
		coll: db.native.Collection(collName),
		name: collName,
	}
}

// Native 提供逃生舱口，允许直接获取原生 *mongo.Collection 进行高级非标操作
func (c *Collection[T]) Native() *mongo.Collection {
	return c.coll
}

// ==========================================
// 核心泛型强类型 CRUD API
// ==========================================

// FindOne 一键式查询单条记录，自动应用拦截器插件，自动反序列化
func (c *Collection[T]) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*T, error) {
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

// Find 一键式查询多条记录，自动应用拦截器插件，自动转换为强类型切片
func (c *Collection[T]) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]T, error) {
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

// InsertOne 写入单条文档，自动执行 BeforeInsert 生命周期拦截（如自动填充 TenantID 与自增 ID）
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

// InsertMany 批量写入文档，自动将 []T 隐式强转，并自动运行批量 ID 审计拦截
func (c *Collection[T]) InsertMany(ctx context.Context, docs []*T, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	stmt := &Statement{
		CollectionName: c.name,
		Model:          &docs, // 传递切片指针，允许插件反射修改每一个元素的值
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

// UpdateOne 更新单条文档，自动拦截并追加严格的防越权隔离
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

// UpdateMany 批量更新文档，自动拦截并追加租户安全条件
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

// DeleteOne 删除单条文档，自动拦截追加写校验隔离
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

// DeleteMany 批量删除文档，自动拦截追加隔离
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
	finalFilter, err := c.applyBeforeFind(ctx, filter, nil)
	if err != nil {
		return 0, err
	}
	return c.coll.CountDocuments(ctx, finalFilter, opts...)
}

// Distinct 获取去重字段集，自动拦截查询
func (c *Collection[T]) Distinct(ctx context.Context, fieldName string, filter interface{}, opts ...*options.DistinctOptions) ([]interface{}, error) {
	finalFilter, err := c.applyBeforeFind(ctx, filter, nil)
	if err != nil {
		return nil, err
	}
	return c.coll.Distinct(ctx, fieldName, finalFilter, opts...)
}

// ==========================================
// 插件生命周期分发辅助引擎
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
