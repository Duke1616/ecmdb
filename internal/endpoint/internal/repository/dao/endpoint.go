package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const EndpointCollection = "c_endpoint"

type EndpointDAO interface {
	// CreateEndpoint 创建单个端点
	CreateEndpoint(ctx context.Context, t Endpoint) (int64, error)
	
	// BatchCreateByResource 按 Resource 批量同步端点
	// 支持智能同步：插入新端点、更新已存在端点、删除不再存在的端点
	BatchCreateByResource(ctx context.Context, resource string, req []Endpoint) (int64, error)
	
	// ListEndpoint 获取端点列表
	ListEndpoint(ctx context.Context, offset, limit int64, path string) ([]Endpoint, error)
	
	// Count 获取端点总数
	Count(ctx context.Context, path string) (int64, error)
}

type endpointDAO struct {
	db *mongox.Mongo
}

// BatchCreateByResource 按 Resource 批量同步端点
func (dao *endpointDAO) BatchCreateByResource(ctx context.Context, resource string, req []Endpoint) (int64, error) {
	if len(req) == 0 {
		return 0, nil
	}

	col := dao.db.Collection(EndpointCollection)
	now := time.Now()

	// 准备 BulkWrite 操作
	var operations []mongo.WriteModel

	// 1. 先获取指定 Resource 的现有端点，用于比较
	existingEndpoints, err := dao.getEndpointsByResourceForSync(ctx, resource)
	if err != nil {
		return 0, fmt.Errorf("获取现有端点失败: %w", err)
	}

	// 2. 创建现有端点的映射，用于快速查找
	existingMap := make(map[string]Endpoint)
	for _, ep := range existingEndpoints {
		key := ep.Method + ":" + ep.Path
		existingMap[key] = ep
	}

	// 3. 处理新端点：不存在的插入，存在的更新
	for i := range req {
		key := req[i].Method + ":" + req[i].Path

		if existingEp, exists := existingMap[key]; exists {
			// 存在且需要更新
			if dao.endpointChanged(existingEp, req[i]) {
				req[i].Id = existingEp.Id
				req[i].Ctime = existingEp.Ctime
				req[i].Utime = now.UnixMilli()

				// 创建更新操作
				filter := bson.M{"_id": existingEp.Id}
				update := bson.M{"$set": req[i]}
				operation := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update)
				operations = append(operations, operation)
			}
			// 从映射中删除，标记为已处理
			delete(existingMap, key)
		} else {
			// 不存在，需要插入
			req[i].Id = dao.db.GetIdGenerator(EndpointCollection)
			req[i].Ctime = now.UnixMilli()
			req[i].Utime = now.UnixMilli()

			// 创建插入操作
			operation := mongo.NewInsertOneModel().SetDocument(req[i])
			operations = append(operations, operation)
		}
	}

	// 4. 删除不再存在的端点（只删除指定 Resource 的）
	for _, ep := range existingMap {
		filter := bson.M{"_id": ep.Id}
		operation := mongo.NewDeleteOneModel().SetFilter(filter)
		operations = append(operations, operation)
	}

	// 5. 执行批量写入
	if len(operations) > 0 {
		opts := options.BulkWrite().SetOrdered(false) // 设置为无序，提高性能
		result, err := col.BulkWrite(ctx, operations, opts)
		if err != nil {
			return 0, fmt.Errorf("批量同步端点失败: %w", err)
		}

		return result.InsertedCount + result.ModifiedCount + result.DeletedCount, nil
	}

	return 0, nil
}

func (dao *endpointDAO) ListEndpoint(ctx context.Context, offset, limit int64, path string) ([]Endpoint, error) {
	col := dao.db.Collection(EndpointCollection)
	filter := bson.M{}
	if path != "" {
		filter = bson.M{"$text": bson.M{"$search": path}}
	}
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	cursor, err := col.Find(ctx, filter, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []Endpoint
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *endpointDAO) Count(ctx context.Context, path string) (int64, error) {
	col := dao.db.Collection(EndpointCollection)
	filter := bson.M{}
	if path != "" {
		filter = bson.M{"$text": bson.M{"$search": path}}
	}
	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *endpointDAO) CreateEndpoint(ctx context.Context, e Endpoint) (int64, error) {
	e.Id = dao.db.GetIdGenerator(EndpointCollection)
	col := dao.db.Collection(EndpointCollection)
	now := time.Now()
	e.Ctime, e.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, e)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return e.Id, nil
}

// 获取所有端点用于同步比较
func (dao *endpointDAO) getAllEndpointsForSync(ctx context.Context) ([]Endpoint, error) {
	col := dao.db.Collection(EndpointCollection)

	cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("查询所有端点失败: %w", err)
	}
	defer cursor.Close(ctx)

	var endpoints []Endpoint
	if err = cursor.All(ctx, &endpoints); err != nil {
		return nil, fmt.Errorf("解码端点数据失败: %w", err)
	}

	return endpoints, nil
}

// 获取指定 Resource 的端点用于同步比较
func (dao *endpointDAO) getEndpointsByResourceForSync(ctx context.Context, resource string) ([]Endpoint, error) {
	col := dao.db.Collection(EndpointCollection)

	filter := bson.M{"resource": resource}
	cursor, err := col.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("查询指定资源端点失败: %w", err)
	}
	defer cursor.Close(ctx)

	var endpoints []Endpoint
	if err = cursor.All(ctx, &endpoints); err != nil {
		return nil, fmt.Errorf("解码端点数据失败: %w", err)
	}

	return endpoints, nil
}

// 检查端点是否发生变化
func (dao *endpointDAO) endpointChanged(existing, new Endpoint) bool {
	// 比较关键字段，忽略 ID 和时间戳
	return existing.Path != new.Path ||
		existing.Method != new.Method ||
		existing.Resource != new.Resource ||
		existing.Desc != new.Desc ||
		existing.IsAuth != new.IsAuth ||
		existing.IsAudit != new.IsAudit ||
		existing.IsPermission != new.IsPermission
}

func NewEndpointDAO(db *mongox.Mongo) EndpointDAO {
	return &endpointDAO{
		db: db,
	}
}

type Endpoint struct {
	Id           int64  `bson:"id"`
	Path         string `bson:"path"`
	Method       string `bson:"method"`
	Resource     string `bson:"resource"`
	Desc         string `bson:"desc"`
	IsAuth       bool   `bson:"is_auth"`
	IsAudit      bool   `bson:"is_audit"`
	IsPermission bool   `bson:"is_permission"`
	Ctime        int64  `bson:"ctime"`
	Utime        int64  `bson:"utime"`
}
