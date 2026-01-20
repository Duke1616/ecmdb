package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const ResourceCollection = "c_resources"

type ResourceDAO interface {
	// CreateResource 创建资产
	CreateResource(ctx context.Context, resource Resource) (int64, error)

	// FindResourceById 根据 ID 和字段列表查询资产
	FindResourceById(ctx context.Context, fields []string, id int64) (Resource, error)

	// ListResource 获取指定模型的资产列表
	ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]Resource, error)

	// CountByModelUid 统计指定模型的资产数量
	CountByModelUid(ctx context.Context, modelUid string) (int64, error)

	// SetCustomField 设置指定资产的自定义字段值
	SetCustomField(ctx context.Context, id int64, field string, data interface{}) (int64, error)

	// ListResourcesByIds 根据 ID 列表批量查询资产
	ListResourcesByIds(ctx context.Context, fields []string, ids []int64) ([]Resource, error)

	// DeleteResource 删除指定资产
	DeleteResource(ctx context.Context, id int64) (int64, error)

	// ListExcludeAndFilterResourceByIds 排除指定 ID 并根据条件过滤资产列表
	ListExcludeAndFilterResourceByIds(ctx context.Context, fields []string, modelUid string, offset, limit int64,
		ids []int64, filter domain.Condition) ([]Resource, error)

	// TotalExcludeAndFilterResourceByIds 排除指定 ID 并根据条件统计资产总数
	TotalExcludeAndFilterResourceByIds(ctx context.Context, modelUid string, ids []int64,
		filter domain.Condition) (int64, error)

	// Search 全局搜索资产
	Search(ctx context.Context, text string) ([]SearchResource, error)

	// FindSecureData 查找指定资产的加密字段数据
	FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error)

	// UpdateAttribute 更新资产属性
	UpdateAttribute(ctx context.Context, resource Resource) (int64, error)

	// CountByModelUids 统计多个模型的资产数量
	CountByModelUids(ctx context.Context, modelUids []string) (map[string]int, error)

	// BatchUpdateResources 批量更新资产
	BatchUpdateResources(ctx context.Context, resources []Resource) (int64, error)

	// BatchCreateOrUpdate 批量创建或更新资产,基于 model_uid + name 进行 upsert
	BatchCreateOrUpdate(ctx context.Context, resources []Resource) error

	// ListBeforeUtime 获取指定时间前的资产列表
	ListBeforeUtime(ctx context.Context, utime int64, fields []string, modelUid string,
		offset, limit int64) ([]Resource, error)
}

type resourceDAO struct {
	db *mongox.Mongo
}

func NewResourceDAO(db *mongox.Mongo) ResourceDAO {
	return &resourceDAO{
		db: db,
	}
}

func (dao *resourceDAO) ListBeforeUtime(ctx context.Context, utime int64, fields []string, modelUid string,
	offset, limit int64) ([]Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"model_uid": modelUid}
	filter["utime"] = bson.M{"$lte": utime}
	projection := buildProjection(fields)
	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
		Sort:       bson.D{{Key: "ctime", Value: -1}},
	}

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误: %w", err)
	}

	var result []Resource
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *resourceDAO) BatchUpdateResources(ctx context.Context, resources []Resource) (int64, error) {
	if len(resources) == 0 {
		return 0, nil
	}

	col := dao.db.Collection(ResourceCollection)
	var totalModified int64

	// 为批量操作创建切片
	models := make([]mongo.WriteModel, 0, len(resources))

	utime := time.Now().UnixMilli()
	for _, resource := range resources {
		updateDoc := bson.M{
			"utime": utime,
		}

		// 将资源数据合并到更新文档中
		for key, value := range resource.Data {
			updateDoc[key] = value
		}

		// 创建更新模型
		filter := bson.M{"id": resource.ID}
		update := bson.M{"$set": updateDoc}

		model := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(false)

		models = append(models, model)
	}

	// 执行批量操作
	result, err := col.BulkWrite(ctx, models)
	if err != nil {
		return 0, fmt.Errorf("批量更新文档操作: %w", err)
	}

	totalModified = result.ModifiedCount
	return totalModified, nil
}

func (dao *resourceDAO) SetCustomField(ctx context.Context, id int64, field string, data interface{}) (int64, error) {
	col := dao.db.Collection(ResourceCollection)

	updateDoc := bson.M{
		"$set": bson.M{
			field:   data,
			"utime": time.Now().UnixMilli(),
		},
	}

	filter := bson.M{"id": id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *resourceDAO) UpdateAttribute(ctx context.Context, resource Resource) (int64, error) {
	col := dao.db.Collection(ResourceCollection)

	updateDoc := bson.M{
		"utime": time.Now().UnixMilli(),
	}

	for key, value := range resource.Data {
		updateDoc[key] = value
	}

	// 构建最终的更新文档
	updateCommand := bson.M{
		"$set": updateDoc,
	}

	filter := bson.M{"id": resource.ID}
	count, err := col.UpdateOne(ctx, filter, updateCommand)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *resourceDAO) FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": id}
	projection := make(map[string]int, 1)
	projection[fieldUid] = 1
	opts := &options.FindOneOptions{
		Projection: projection,
	}

	var result = make(map[string]string)

	if err := col.FindOne(ctx, filter, opts).Decode(&result); err != nil {
		return "", fmt.Errorf("解码错误: %w", err)
	}

	fieldValue, ok := result[fieldUid]

	if !ok {
		return "无", nil
	}

	return fieldValue, nil
}

func (dao *resourceDAO) CreateResource(ctx context.Context, r Resource) (int64, error) {
	now := time.Now()
	r.Ctime, r.Utime = now.UnixMilli(), now.UnixMilli()
	r.ID = dao.db.GetIdGenerator(ResourceCollection)
	col := dao.db.Collection(ResourceCollection)

	_, err := col.InsertOne(ctx, r)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return r.ID, nil
}

func (dao *resourceDAO) FindResourceById(ctx context.Context, fields []string, id int64) (Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": id}
	projection := buildProjection(fields)
	opts := &options.FindOneOptions{
		Projection: projection,
	}

	var result Resource
	if err := col.FindOne(ctx, filter, opts).Decode(&result); err != nil {
		return Resource{}, fmt.Errorf("解码错误: %w", err)
	}
	return result, nil
}

func (dao *resourceDAO) ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"model_uid": modelUid}
	projection := buildProjection(fields)
	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
		Sort:       bson.D{{Key: "ctime", Value: -1}},
	}

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误: %w", err)
	}

	var result []Resource
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *resourceDAO) CountByModelUid(ctx context.Context, modelUid string) (int64, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"model_uid": modelUid}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *resourceDAO) ListResourcesByIds(ctx context.Context, fields []string, ids []int64) ([]Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": bson.M{"$in": ids}}
	projection := make(map[string]int, len(fields))
	for _, v := range fields {
		projection[v] = 1
	}
	projection["_id"] = 0
	projection["id"] = 1
	projection["name"] = 1
	projection["model_uid"] = 1
	opts := &options.FindOptions{
		Projection: projection,
	}

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误: %w", err)
	}

	var result []Resource
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *resourceDAO) DeleteResource(ctx context.Context, id int64) (int64, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"id": id}

	result, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *resourceDAO) CountByModelUids(ctx context.Context, modelUids []string) (map[string]int, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{}
	if len(modelUids) > 0 {
		filter["model_uid"] = bson.D{{Key: "$in", Value: modelUids}}
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$model_uid"},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	// 执行聚合查询
	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}
	defer cursor.Close(ctx)

	// 定义结果结构体
	var results []struct {
		ModelUID string `bson:"_id"`
		Total    int    `bson:"total"`
	}

	// 将游标中的数据解码到 results 变量
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}

	// 将结果转换为 map[model_uid]total_count
	modelCountMap := make(map[string]int)
	for _, result := range results {
		modelCountMap[result.ModelUID] = result.Total
	}

	return modelCountMap, nil

}

func (dao *resourceDAO) Search(ctx context.Context, text string) ([]SearchResource, error) {
	col := dao.db.Collection(ResourceCollection)
	filter := bson.M{"$text": bson.M{"$search": text}}
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$model_uid"},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
		}},
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		groupStage,
		{{Key: "$sort", Value: bson.D{{Key: "total", Value: -1}}}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}
	defer cursor.Close(ctx)

	var result []SearchResource
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}

	return result, nil
}

func (dao *resourceDAO) ListExcludeAndFilterResourceByIds(ctx context.Context, fields []string, modelUid string,
	offset, limit int64, ids []int64, filter domain.Condition) ([]Resource, error) {
	col := dao.db.Collection(ResourceCollection)
	filters := bson.M{"model_uid": modelUid}
	if len(ids) > 0 {
		filters["id"] = bson.M{
			"$nin": ids,
		}
	}

	switch filter.Condition {
	case "not_equal":
		filters[filter.Name] = bson.M{"$ne": filter.Input}
	case "equal":
		filters[filter.Name] = filter.Input
	case "contains":
		filters[filter.Name] = bson.M{"$regex": primitive.Regex{Pattern: filter.Input, Options: "i"}}
	}

	projection := buildProjection(fields)

	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
	}

	cursor, err := col.Find(ctx, filters, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误: %w", err)
	}

	var result []Resource
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *resourceDAO) TotalExcludeAndFilterResourceByIds(ctx context.Context, modelUid string,
	ids []int64, filter domain.Condition) (int64, error) {
	col := dao.db.Collection(ResourceCollection)
	filters := bson.M{"model_uid": modelUid}
	if len(ids) > 0 {
		filters["id"] = bson.M{
			"$nin": ids,
		}
	}

	switch filter.Condition {
	case "not_equal":
		filters[filter.Name] = bson.M{"$ne": filter.Input}
	case "equal":
		filters[filter.Name] = filter.Input
	case "contains":
		filters[filter.Name] = bson.M{"$regex": primitive.Regex{Pattern: filter.Input, Options: "i"}}
	}

	count, err := col.CountDocuments(ctx, filters)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

type Resource struct {
	ID       int64         `bson:"id"`
	ModelUID string        `bson:"model_uid"`
	Data     mongox.MapStr `bson:",inline"`
	Ctime    int64         `bson:"ctime"`
	Utime    int64         `bson:"utime"`
}

type Pipeline struct {
	ModelUid string `bson:"_id"`
	Total    int    `bson:"total"`
}

type SearchResource struct {
	ModelUid string          `bson:"_id"`
	Total    int             `bson:"total"`
	Data     []mongox.MapStr `bson:"data"`
}

// BatchCreateOrUpdate 批量创建或更新资产
// NOTE: 基于 model_uid + name 进行 upsert,使用 MongoDB BulkWrite 提升性能
func (dao *resourceDAO) BatchCreateOrUpdate(ctx context.Context, resources []Resource) error {
	if len(resources) == 0 {
		return nil
	}

	col := dao.db.Collection(ResourceCollection)
	now := time.Now().UnixMilli()

	// 批量获取 ID(用于新创建的资源)
	// NOTE: 如果是更新操作,$setOnInsert 不会生效,ID 会被浪费,但这不是问题
	startID, err := dao.db.GetBatchIdGenerator(ResourceCollection, len(resources))
	if err != nil {
		return fmt.Errorf("获取批量 ID 失败: %w", err)
	}

	// 构建 BulkWrite 模型
	models := make([]mongo.WriteModel, 0, len(resources))
	currentID := startID

	for _, resource := range resources {
		// 构建 filter: 基于 model_uid + name
		filter := bson.M{
			"model_uid": resource.ModelUID,
			"data.name": resource.Data["name"],
		}

		// 构建 update 文档
		updateDoc := bson.M{
			"utime": now,
		}

		// 合并资源数据
		for key, value := range resource.Data {
			updateDoc[key] = value
		}

		// 构建 upsert 操作
		update := bson.M{
			"$set": updateDoc,
			"$setOnInsert": bson.M{
				"id":    currentID,
				"ctime": now,
			},
		}

		model := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)

		models = append(models, model)
		currentID++
	}

	// 执行批量操作
	_, err = col.BulkWrite(ctx, models)
	if err != nil {
		return fmt.Errorf("批量创建或更新操作失败: %w", err)
	}

	return nil
}
