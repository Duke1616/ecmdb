package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RelationResourceDAO interface {
	CreateResourceRelation(ctx context.Context, mr ResourceRelation) (int64, error)

	ListSrcResources(ctx context.Context, modelUid string, id int64) ([]ResourceRelation, error)
	ListDstResources(ctx context.Context, modelUid string, id int64) ([]ResourceRelation, error)
	CountSrc(ctx context.Context, modelUid string, id int64) (int64, error)
	CountDst(ctx context.Context, modelUid string, id int64) (int64, error)

	ListSrcAggregated(ctx context.Context, modelUid string, id int64) ([]ResourceAggregatedAsset, error)
	ListDstAggregated(ctx context.Context, modelUid string, id int64) ([]ResourceAggregatedAsset, error)

	ListSrcRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error)
	ListDstRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error)

	DeleteResourceRelation(ctx context.Context, id int64) (int64, error)
	DeleteSrcRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error)
	DeleteDstRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error)

	// CountByRelationTypeUid 根据关联类型 UID 获取数量
	CountByRelationTypeUid(ctx context.Context, uid string) (int64, error)

	// CountByRelationName 根据关联名称获取数量
	CountByRelationName(ctx context.Context, name string) (int64, error)
}

func NewRelationResourceDAO(db *mongox.DB) RelationResourceDAO {
	return &resourceDAO{
		db:   db,
		coll: mongox.NewCollection[ResourceRelation](db, ResourceRelationCollection),
	}
}

type resourceDAO struct {
	db   *mongox.DB
	coll *mongox.Collection[ResourceRelation]
}

func (dao *resourceDAO) CreateResourceRelation(ctx context.Context, rr ResourceRelation) (int64, error) {
	now := time.Now()
	rr.Ctime, rr.Utime = now.UnixMilli(), now.UnixMilli()

	// 借助 AutoIDPlugin 插件自动分配自增主键，无须再手动管理自增 ID。
	_, err := dao.coll.InsertOne(ctx, &rr)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return rr.Id, nil
}

func (dao *resourceDAO) ListSrcResources(ctx context.Context, modelUid string, id int64) ([]ResourceRelation, error) {
	filter := bson.M{
		"source_model_uid":   modelUid,
		"source_resource_id": id,
	}
	opts := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	return dao.coll.Find(ctx, filter, opts)
}

func (dao *resourceDAO) ListDstResources(ctx context.Context, modelUid string, id int64) ([]ResourceRelation, error) {
	filter := bson.M{
		"target_model_uid":   modelUid,
		"target_resource_id": id,
	}
	opts := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	return dao.coll.Find(ctx, filter, opts)
}

func (dao *resourceDAO) CountSrc(ctx context.Context, modelUid string, id int64) (int64, error) {
	filter := bson.M{
		"source_model_uid":   modelUid,
		"source_resource_id": id,
	}

	count, err := dao.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *resourceDAO) CountDst(ctx context.Context, modelUid string, id int64) (int64, error) {
	filter := bson.M{
		"target_model_uid":   modelUid,
		"target_resource_id": id,
	}

	count, err := dao.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *resourceDAO) ListSrcAggregated(ctx context.Context, modelUid string, id int64) ([]ResourceAggregatedAsset, error) {
	filter := bson.M{
		"source_model_uid":   modelUid,
		"source_resource_id": id,
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$relation_name"},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}},                             // 统计每个分组中的文档数量
			{Key: "resource_ids", Value: bson.D{{Key: "$push", Value: "$target_resource_id"}}}, // 将目标资源 Ids 添加到一个数组中
			{Key: "model_uid", Value: bson.D{{Key: "$first", Value: "$target_model_uid"}}},     // 添加额外字段
		}}},
	}

	cursor, err := dao.coll.Native().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}
	defer cursor.Close(ctx)

	var result []ResourceAggregatedAsset
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}

	return result, nil
}

func (dao *resourceDAO) ListDstAggregated(ctx context.Context, modelUid string, id int64) ([]ResourceAggregatedAsset, error) {
	filter := bson.M{
		"target_model_uid":   modelUid,
		"target_resource_id": id,
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}}, // 添加筛选条件
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$relation_name"},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}},                             // 统计每个分组中的文档数量
			{Key: "resource_ids", Value: bson.D{{Key: "$push", Value: "$source_resource_id"}}}, // 将源资源 Ids 添加到一个数组中
			{Key: "model_uid", Value: bson.D{{Key: "$first", Value: "$source_model_uid"}}},     // 添加额外字段
		}}},
	}

	cursor, err := dao.coll.Native().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}
	defer cursor.Close(ctx)

	var result []ResourceAggregatedAsset
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}

	return result, nil
}

func (dao *resourceDAO) ListSrcRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error) {
	filter := bson.M{
		"source_model_uid":   modelUid,
		"relation_name":      relationName,
		"source_resource_id": id,
	}

	opts := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	results, err := dao.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	// NOTE: 借助 lo.Map 消除繁琐的手动游标 Decode 循环与内部变量声明
	return lo.Map(results, func(rr ResourceRelation, _ int) int64 {
		return rr.TargetResourceID
	}), nil
}

func (dao *resourceDAO) ListDstRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error) {
	filter := bson.M{
		"target_model_uid":   modelUid,
		"relation_name":      relationName,
		"target_resource_id": id,
	}
	opts := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	results, err := dao.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	// NOTE: 借助 lo.Map 消除繁琐的手动游标 Decode 循环与内部变量声明
	return lo.Map(results, func(rr ResourceRelation, _ int) int64 {
		return rr.SourceResourceID
	}), nil
}

func (dao *resourceDAO) DeleteResourceRelation(ctx context.Context, id int64) (int64, error) {
	filter := bson.M{"id": id}

	result, err := dao.coll.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *resourceDAO) DeleteSrcRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error) {
	filter := bson.M{
		"source_model_uid":   modelUid,
		"source_resource_id": resourceId,
		"relation_name":      relationName,
	}

	result, err := dao.coll.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *resourceDAO) DeleteDstRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error) {
	filter := bson.M{
		"target_model_uid":   modelUid,
		"target_resource_id": resourceId,
		"relation_name":      relationName,
	}

	result, err := dao.coll.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *resourceDAO) CountByRelationTypeUid(ctx context.Context, uid string) (int64, error) {
	filter := bson.M{"relation_type_uid": uid}
	count, err := dao.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("关联引用统计错误: %w", err)
	}
	return count, nil
}

func (dao *resourceDAO) CountByRelationName(ctx context.Context, name string) (int64, error) {
	filter := bson.M{"relation_name": name}
	count, err := dao.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("关联引用统计错误: %w", err)
	}
	return count, nil
}

type ResourceRelation struct {
	Id               int64  `bson:"id"`
	SourceModelUID   string `bson:"source_model_uid"`
	TargetModelUID   string `bson:"target_model_uid"`
	SourceResourceID int64  `bson:"source_resource_id"`
	TargetResourceID int64  `bson:"target_resource_id"`
	RelationTypeUID  string `bson:"relation_type_uid"`
	RelationName     string `bson:"relation_name"`
	Ctime            int64  `bson:"ctime"`
	Utime            int64  `bson:"utime"`
}

func (a *ResourceRelation) SetID(id int64) {
	a.Id = id
}

func (a *ResourceRelation) GetID() int64 {
	return a.Id
}

// ResourceAggregatedAsset 聚合查询返回数据
type ResourceAggregatedAsset struct {
	RelationName string  `bson:"_id"`
	ModelUid     string  `bson:"model_uid"`
	Total        int     `bson:"total"`
	ResourceIds  []int64 `bson:"resource_ids"`
}
