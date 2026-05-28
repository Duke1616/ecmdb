package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/Duke1616/eiam/pkg/ctxutil"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
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

	// UnsetCustomField 抹除指定模型下所有资产的自定义字段（平铺键）
	UnsetCustomField(ctx context.Context, modelUid string, fieldUid string) (int64, error)

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

	// ListResourcesWithFilters 根据复杂筛选条件获取资产列表
	ListResourcesWithFilters(ctx context.Context, fields []string, modelUid string, ids []int64, offset, limit int64,
		filterGroups []domain.FilterGroup) ([]Resource, error)

	// TotalResourcesWithFilters 根据复杂筛选条件统计资产数量
	TotalResourcesWithFilters(ctx context.Context, modelUid string, ids []int64, filterGroups []domain.FilterGroup) (int64, error)
}

type resourceDAO struct {
	db   *mongox.DB
	coll *mongox.Collection[Resource]
}

func NewResourceDAO(db *mongox.DB) ResourceDAO {
	return &resourceDAO{
		db:   db,
		coll: mongox.NewCollection[Resource](db, ResourceCollection),
	}
}

func (dao *resourceDAO) ListBeforeUtime(ctx context.Context, utime int64, fields []string, modelUid string,
	offset, limit int64) ([]Resource, error) {
	filter := bson.M{"model_uid": modelUid}
	filter["utime"] = bson.M{"$lte": utime}
	projection := buildProjection(fields)
	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
		Sort:       bson.D{{Key: "ctime", Value: -1}},
	}

	return dao.coll.Find(ctx, filter, opts)
}

func (dao *resourceDAO) BatchUpdateResources(ctx context.Context, resources []Resource) (int64, error) {
	if len(resources) == 0 {
		return 0, nil
	}

	utime := time.Now().UnixMilli()
	models := lo.Map(resources, func(r Resource, _ int) mongo.WriteModel {
		return mongo.NewUpdateOneModel().
			SetFilter(bson.M{"id": r.ID}).
			SetUpdate(bson.M{"$set": dao.buildUpdateDoc(r.Data, utime)}).
			SetUpsert(false)
	})

	result, err := dao.coll.BulkWrite(ctx, models)
	if err != nil {
		return 0, fmt.Errorf("批量更新文档操作: %w", err)
	}

	return result.ModifiedCount, nil
}

func (dao *resourceDAO) SetCustomField(ctx context.Context, id int64, field string, data interface{}) (int64, error) {
	updateDoc := bson.M{
		"$set": bson.M{
			field:   data,
			"utime": time.Now().UnixMilli(),
		},
	}

	filter := bson.M{"id": id}
	count, err := dao.coll.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *resourceDAO) UpdateAttribute(ctx context.Context, resource Resource) (int64, error) {
	updateCommand := bson.M{
		"$set": dao.buildUpdateDoc(resource.Data, time.Now().UnixMilli()),
	}

	filter := bson.M{"id": resource.ID}
	count, err := dao.coll.UpdateOne(ctx, filter, updateCommand)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *resourceDAO) FindSecureData(ctx context.Context, id int64, fieldUid string) (string, error) {
	filter := bson.M{"id": id}
	projection := make(map[string]int, 1)
	projection[fieldUid] = 1
	opts := &options.FindOneOptions{
		Projection: projection,
	}

	var result = make(map[string]string)
	if err := dao.coll.Native().FindOne(ctx, filter, opts).Decode(&result); err != nil {
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

	// 依靠 mongox 的 AutoIDPlugin 插件自动分配并注入 ID，不需要再手动管理 id_generator
	_, err := dao.coll.InsertOne(ctx, &r)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return r.ID, nil
}

func (dao *resourceDAO) FindResourceById(ctx context.Context, fields []string, id int64) (Resource, error) {
	filter := bson.M{"id": id}
	projection := buildProjection(fields)
	opts := &options.FindOneOptions{
		Projection: projection,
	}

	m, err := dao.coll.FindOne(ctx, filter, opts)
	if err != nil {
		return Resource{}, fmt.Errorf("解码错误: %w", err)
	}
	return *m, nil
}

func (dao *resourceDAO) ListResource(ctx context.Context, fields []string, modelUid string, offset, limit int64) ([]Resource, error) {
	filter := bson.M{"model_uid": modelUid}
	projection := buildProjection(fields)
	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
		Sort:       bson.D{{Key: "ctime", Value: -1}},
	}

	return dao.coll.Find(ctx, filter, opts)
}

func (dao *resourceDAO) CountByModelUid(ctx context.Context, modelUid string) (int64, error) {
	filter := bson.M{"model_uid": modelUid}
	count, err := dao.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *resourceDAO) ListResourcesByIds(ctx context.Context, fields []string, ids []int64) ([]Resource, error) {
	filter := bson.M{"id": bson.M{"$in": ids}}
	projection := buildProjection(fields)
	opts := &options.FindOptions{
		Projection: projection,
	}

	return dao.coll.Find(ctx, filter, opts)
}

func (dao *resourceDAO) DeleteResource(ctx context.Context, id int64) (int64, error) {
	filter := bson.M{"id": id}
	result, err := dao.coll.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *resourceDAO) CountByModelUids(ctx context.Context, modelUids []string) (map[string]int, error) {
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

	cursor, err := dao.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		ModelUID string `bson:"_id"`
		Total    int    `bson:"total"`
	}

	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}

	// NOTE: 使用 lo.Associate 将结构体切片直接映射并组装为计数 map
	modelCountMap := lo.Associate(results, func(r struct {
		ModelUID string `bson:"_id"`
		Total    int    `bson:"total"`
	}) (string, int) {
		return r.ModelUID, r.Total
	})

	return modelCountMap, nil
}

func (dao *resourceDAO) Search(ctx context.Context, text string) ([]SearchResource, error) {
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
		{{Key: "$limit", Value: 1000}}, // NOTE: 极限防御：限制匹配上限，阻断全文检索匹配数万文档触发 16MB 崩溃与内存溢出灾难
		groupStage,
		{{Key: "$sort", Value: bson.D{{Key: "total", Value: -1}}}},
	}

	cursor, err := dao.coll.Aggregate(ctx, pipeline)
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
	filters := dao.buildExcludeAndFilterBson(modelUid, ids, filter)
	projection := buildProjection(fields)
	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
	}

	return dao.coll.Find(ctx, filters, opts)
}

func (dao *resourceDAO) TotalExcludeAndFilterResourceByIds(ctx context.Context, modelUid string,
	ids []int64, filter domain.Condition) (int64, error) {
	filters := dao.buildExcludeAndFilterBson(modelUid, ids, filter)
	count, err := dao.coll.CountDocuments(ctx, filters)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

// BatchCreateOrUpdate 批量创建或更新资产
func (dao *resourceDAO) BatchCreateOrUpdate(ctx context.Context, resources []Resource) error {
	if len(resources) == 0 {
		return nil
	}

	now := time.Now().UnixMilli()
	tenantID := ctxutil.GetTenantID(ctx).Int64()

	// NOTE: 提取唯一键生成逻辑，消除多处重复的 fmt.Sprintf 格式化散落
	resourceKey := func(modelUID string, data mongox.MapStr) string {
		name, _ := data["name"].(string)
		return modelUID + "_" + name
	}

	// 1. 构建 Index-Covered 覆盖索引批量查询条件（框架层 Find 会自动追加 tenant_id 过滤）
	filters := lo.Map(resources, func(r Resource, _ int) interface{} {
		return bson.M{
			"model_uid": r.ModelUID,
			"data.name": r.Data["name"],
		}
	})

	existingDocs, err := dao.coll.Find(ctx, bson.M{"$or": filters}, &options.FindOptions{
		Projection: bson.M{"model_uid": 1, "data.name": 1, "_id": 0},
	})
	if err != nil {
		return fmt.Errorf("批量查询已存在资产失败: %w", err)
	}

	// 2. 构建已存在资产的唯一键集合
	existingMap := lo.Associate(existingDocs, func(doc Resource) (string, struct{}) {
		return resourceKey(doc.ModelUID, doc.Data), struct{}{}
	})

	// 3. 精准统计真正需要执行 Insert 的新资产数量
	needInsertCount := lo.CountBy(resources, func(r Resource) bool {
		_, exists := existingMap[resourceKey(r.ModelUID, r.Data)]
		return !exists
	})

	// 4. 按需、精准申请自增 ID，消灭序列号空洞与原子锁冲突
	var startID int64
	if needInsertCount > 0 {
		startID, err = dao.db.GetBatchIdGenerator(ResourceCollection, needInsertCount)
		if err != nil {
			return fmt.Errorf("获取批量 ID 失败: %w", err)
		}
	}

	// 5. 动态分配 ID 并组装 BulkWrite 写入模型（框架层 BulkWrite 会自动为 Filter 追加 tenant_id 过滤拦截）
	var insertIndex int64
	models := lo.Map(resources, func(r Resource, _ int) mongo.WriteModel {
		key := resourceKey(r.ModelUID, r.Data)

		var docID int64
		if _, exists := existingMap[key]; !exists {
			docID = startID + insertIndex
			insertIndex++
		}

		return mongo.NewUpdateOneModel().
			SetFilter(bson.M{
				"model_uid": r.ModelUID,
				"data.name": r.Data["name"],
			}).
			SetUpdate(bson.M{
				"$set": dao.buildUpdateDoc(r.Data, now),
				"$setOnInsert": bson.M{
					"id":        docID,
					"tenant_id": tenantID, // 为防范 MongoDB Upsert 复合条件提取失效，此处仍显式保留 tenant_id 写入以确保新文档物理归属
					"ctime":     now,
				},
			}).
			SetUpsert(true)
	})

	_, err = dao.coll.BulkWrite(ctx, models)
	if err != nil {
		return fmt.Errorf("批量创建或更新操作失败: %w", err)
	}

	return nil
}

func (dao *resourceDAO) ListResourcesWithFilters(ctx context.Context, fields []string, modelUid string, ids []int64, offset, limit int64, filterGroups []domain.FilterGroup) ([]Resource, error) {
	baseFilter := bson.M{"model_uid": modelUid}
	if len(ids) > 0 {
		baseFilter["id"] = bson.M{"$in": ids}
	}

	if len(filterGroups) == 0 {
		projection := buildProjection(fields)
		opts := &options.FindOptions{
			Projection: projection,
			Limit:      &limit,
			Skip:       &offset,
			Sort:       bson.D{{Key: "ctime", Value: -1}},
		}
		return dao.coll.Find(ctx, baseFilter, opts)
	}

	// NOTE: 统一调用内部提炼的 buildFilterConditions 辅助拼装器，消灭 18 行冗余 Duplicate 条件树生成逻辑
	orConditions := buildFilterConditions(filterGroups)
	finalFilter := dao.combineFilters(baseFilter, orConditions)

	projection := buildProjection(fields)
	opts := &options.FindOptions{
		Projection: projection,
		Limit:      &limit,
		Skip:       &offset,
		Sort:       bson.D{{Key: "ctime", Value: -1}},
	}

	return dao.coll.Find(ctx, finalFilter, opts)
}

func (dao *resourceDAO) TotalResourcesWithFilters(ctx context.Context, modelUid string, ids []int64, filterGroups []domain.FilterGroup) (int64, error) {
	baseFilter := bson.M{"model_uid": modelUid}
	if len(ids) > 0 {
		baseFilter["id"] = bson.M{"$in": ids}
	}

	if len(filterGroups) == 0 {
		return dao.coll.CountDocuments(ctx, baseFilter)
	}

	// NOTE: 统一调用内部提炼的 buildFilterConditions 辅助拼装器，消灭 18 行冗余 Duplicate 条件树生成逻辑
	orConditions := buildFilterConditions(filterGroups)
	finalFilter := dao.combineFilters(baseFilter, orConditions)

	count, err := dao.coll.CountDocuments(ctx, finalFilter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}
	return count, nil
}

type Resource struct {
	TenantID int64         `bson:"tenant_id"`
	ID       int64         `bson:"id"`
	ModelUID string        `bson:"model_uid"`
	Data     mongox.MapStr `bson:",inline"`
	Ctime    int64         `bson:"ctime"`
	Utime    int64         `bson:"utime"`
}

func (r *Resource) SetID(id int64) {
	r.ID = id
}

func (r *Resource) GetID() int64 {
	return r.ID
}

func (dao *resourceDAO) UnsetCustomField(ctx context.Context, modelUid string, fieldUid string) (int64, error) {
	filter := bson.M{"model_uid": modelUid}
	update := bson.M{"$unset": bson.M{fieldUid: ""}}

	result, err := dao.coll.Native().UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("批量抹除平铺字段错误: %w", err)
	}

	return result.ModifiedCount, nil
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
