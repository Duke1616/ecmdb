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
	CreateResource(ctx context.Context, resource Resource) (int64, error)
	FindResourceById(ctx context.Context, fields []string, id int64) (Resource, error)
	ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]Resource, error)
	CountByModelUid(ctx context.Context, modelUid string) (int64, error)
	SetCustomField(ctx context.Context, id int64, field string, data interface{}) (int64, error)
	ListResourcesByIds(ctx context.Context, fields []string, ids []int64) ([]Resource, error)
	DeleteResource(ctx context.Context, id int64) (int64, error)
	ListExcludeAndFilterResourceByIds(ctx context.Context, fields []string, modelUid string, offset, limit int64,
		ids []int64, filter domain.Condition) ([]Resource, error)
	TotalExcludeAndFilterResourceByIds(ctx context.Context, modelUid string, ids []int64,
		filter domain.Condition) (int64, error)
	Search(ctx context.Context, text string) ([]SearchResource, error)

	FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error)
	UpdateAttribute(ctx context.Context, resource Resource) (int64, error)

	CountByModelUids(ctx context.Context, modelUids []string) (map[string]int, error)
}

type resourceDAO struct {
	db *mongox.Mongo
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

func NewResourceDAO(db *mongox.Mongo) ResourceDAO {
	return &resourceDAO{
		db: db,
	}
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
	projection := make(map[string]int, len(fields))
	for _, v := range fields {
		projection[v] = 1
	}
	projection["_id"] = 0
	projection["id"] = 1
	projection["name"] = 1
	projection["model_uid"] = 1
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
	projection := make(map[string]int, len(fields))
	for _, v := range fields {
		projection[v] = 1
	}
	projection["_id"] = 0
	projection["id"] = 1
	projection["name"] = 1

	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
		Sort:       bson.D{{Key: "ctime", Value: -1}},
	}

	cursor, err := col.Find(ctx, filter, opts)
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
		filter["model_uid"] = bson.D{{"$in", modelUids}}
	}

	pipeline := mongo.Pipeline{
		{{"$match", filter}},
		{{"$group", bson.D{
			{"_id", "$model_uid"},
			{"total", bson.D{{"$sum", 1}}},
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
		{"$group", bson.D{
			{"_id", "$model_uid"},
			{"total", bson.D{{"$sum", 1}}},
			{"data", bson.D{{"$push", "$$ROOT"}}},
		}},
	}

	pipeline := mongo.Pipeline{
		{{"$match", filter}},
		groupStage,
		{{"$sort", bson.D{{"total", -1}}}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

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

	projection := make(map[string]int, len(fields))
	for _, v := range fields {
		projection[v] = 1
	}
	projection["_id"] = 0
	projection["id"] = 1
	projection["model_uid"] = 1
	projection["name"] = 1

	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
	}

	cursor, err := col.Find(ctx, filters, opts)
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
