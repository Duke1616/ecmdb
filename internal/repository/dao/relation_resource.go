package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/Duke1616/eiam/pkg/ctxutil"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ResourceRelationCollection = "c_relation_resource"
)

// RelationResourceDAO 资产实例关联关系底层数据访问接口
type RelationResourceDAO interface {
	// CreateResourceRelation 创建单条资产实例关联关系
	CreateResourceRelation(ctx context.Context, mr ResourceRelation) (int64, error)

	// ListSrcResources 查询以指定资产作为源端的正向直接关联列表
	ListSrcResources(ctx context.Context, modelUid string, id int64) ([]ResourceRelation, error)

	// ListDstResources 查询以指定资产作为目标端的反向直接关联列表
	ListDstResources(ctx context.Context, modelUid string, id int64) ([]ResourceRelation, error)

	// CountSrc 统计以指定资产作为源端的正向关联关系总数
	CountSrc(ctx context.Context, modelUid string, id int64) (int64, error)

	// CountDst 统计以指定资产作为目标端的反向关联关系总数
	CountDst(ctx context.Context, modelUid string, id int64) (int64, error)

	// ListSrcAggregated 聚合统计源端资产在各关联关系类型下的目标资产 ID 集合与数量
	ListSrcAggregated(ctx context.Context, modelUid string, id int64) ([]ResourceAggregatedAsset, error)

	// ListDstAggregated 聚合统计目标端资产在各关联关系类型下的源端资产 ID 集合与数量
	ListDstAggregated(ctx context.Context, modelUid string, id int64) ([]ResourceAggregatedAsset, error)

	// ListSrcRelated 查询指定源资产在特定关系类型下已关联的目标资产 ID 列表（用于新增关联时的防重过滤）
	ListSrcRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error)

	// ListDstRelated 查询指定目标资产在特定关系类型下已关联的源资产 ID 列表（用于新增关联时的防重过滤）
	ListDstRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error)

	// DeleteResourceRelation 根据关系主键 ID 删除单条关联关系
	DeleteResourceRelation(ctx context.Context, id int64) (int64, error)

	// DeleteSrcRelation 根据源端资产信息和关系名称，解除源资产的正向关联
	DeleteSrcRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error)

	// DeleteDstRelation 根据目标端资产信息和关系名称，解除目标资产的反向关联
	DeleteDstRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error)

	// DeleteByResourceId 级联删除指定资产关联的所有连边关系（无论作为源端还是目标端，彻底杜绝孤儿悬空节点）
	DeleteByResourceId(ctx context.Context, resourceId int64) (int64, error)

	// CountByRelationTypeUid 统计引用了指定关联类型 UID 的资产关系总数（用于模型元数据删除保护校验）
	CountByRelationTypeUid(ctx context.Context, uid string) (int64, error)

	// CountByRelationName 统计引用了指定模型关联唯一标识的资产关系总数（用于关系模型删除保护校验）
	CountByRelationName(ctx context.Context, name string) (int64, error)

	// ListRecursiveSrc 利用 MongoDB $graphLookup 递归遍历下游关联资产（正向拓扑树生成）
	ListRecursiveSrc(ctx context.Context, modelUid string, id int64, maxDepth int) ([]ResourceRelation, error)

	// ListRecursiveDst 利用 MongoDB $graphLookup 递归遍历上游关联资产（反向拓扑树生成）
	ListRecursiveDst(ctx context.Context, modelUid string, id int64, maxDepth int) ([]ResourceRelation, error)
}

func NewRelationResourceDAO(db *mongox.DB) RelationResourceDAO {
	return &resourceRelationDAO{
		db:   db,
		coll: mongox.NewCollection[ResourceRelation](db, ResourceRelationCollection),
	}
}

type resourceRelationDAO struct {
	db   *mongox.DB
	coll *mongox.Collection[ResourceRelation]
}

func (dao *resourceRelationDAO) CreateResourceRelation(ctx context.Context, rr ResourceRelation) (int64, error) {
	now := time.Now()
	rr.Ctime, rr.Utime = now.UnixMilli(), now.UnixMilli()

	// 借助 AutoIDPlugin 插件自动分配自增主键，无须再手动管理自增 ID。
	_, err := dao.coll.InsertOne(ctx, &rr)
	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return 0, fmt.Errorf("资产关联关系已存在，请勿重复创建: %w", errs.ErrUniqueDuplicate)
		}
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return rr.Id, nil
}

func (dao *resourceRelationDAO) ListSrcResources(ctx context.Context, modelUid string, id int64) ([]ResourceRelation, error) {
	filter := bson.M{
		"source_model_uid":   modelUid,
		"source_resource_id": id,
	}
	opts := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	return dao.coll.Find(ctx, filter, opts)
}

func (dao *resourceRelationDAO) ListDstResources(ctx context.Context, modelUid string, id int64) ([]ResourceRelation, error) {
	filter := bson.M{
		"target_model_uid":   modelUid,
		"target_resource_id": id,
	}
	opts := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	return dao.coll.Find(ctx, filter, opts)
}

func (dao *resourceRelationDAO) CountSrc(ctx context.Context, modelUid string, id int64) (int64, error) {
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

func (dao *resourceRelationDAO) CountDst(ctx context.Context, modelUid string, id int64) (int64, error) {
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

func (dao *resourceRelationDAO) aggregateRelations(ctx context.Context, modelField, idField, pushID, peerModel string, modelUid string, id int64) ([]ResourceAggregatedAsset, error) {
	filter := bson.M{
		modelField: modelUid,
		idField:    id,
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$relation_name"},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "resource_ids", Value: bson.D{{Key: "$push", Value: "$" + pushID}}},
			{Key: "model_uid", Value: bson.D{{Key: "$first", Value: "$" + peerModel}}},
		}}},
	}

	cursor, err := dao.coll.Aggregate(ctx, pipeline)
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

func (dao *resourceRelationDAO) ListSrcAggregated(ctx context.Context, modelUid string, id int64) ([]ResourceAggregatedAsset, error) {
	return dao.aggregateRelations(ctx, "source_model_uid", "source_resource_id", "target_resource_id", "target_model_uid", modelUid, id)
}

func (dao *resourceRelationDAO) ListDstAggregated(ctx context.Context, modelUid string, id int64) ([]ResourceAggregatedAsset, error) {
	return dao.aggregateRelations(ctx, "target_model_uid", "target_resource_id", "source_resource_id", "source_model_uid", modelUid, id)
}

func (dao *resourceRelationDAO) ListSrcRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error) {
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

func (dao *resourceRelationDAO) ListDstRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error) {
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

func (dao *resourceRelationDAO) DeleteResourceRelation(ctx context.Context, id int64) (int64, error) {
	filter := bson.M{"id": id}

	result, err := dao.coll.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *resourceRelationDAO) DeleteSrcRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error) {
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

func (dao *resourceRelationDAO) DeleteDstRelation(ctx context.Context, resourceId int64, modelUid, relationName string) (int64, error) {
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

func (dao *resourceRelationDAO) DeleteByResourceId(ctx context.Context, resourceId int64) (int64, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"source_resource_id": resourceId},
			{"target_resource_id": resourceId},
		},
	}

	result, err := dao.coll.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("级联删除关联关系错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *resourceRelationDAO) CountByRelationTypeUid(ctx context.Context, uid string) (int64, error) {
	filter := bson.M{"relation_type_uid": uid}
	count, err := dao.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("关联引用统计错误: %w", err)
	}
	return count, nil
}

func (dao *resourceRelationDAO) CountByRelationName(ctx context.Context, name string) (int64, error) {
	filter := bson.M{"relation_name": name}
	count, err := dao.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("关联引用统计错误: %w", err)
	}
	return count, nil
}

func (dao *resourceRelationDAO) ListRecursiveSrc(ctx context.Context, modelUid string, id int64, maxDepth int) ([]ResourceRelation, error) {
	tenantID := ctxutil.GetTenantID(ctx).Int64()

	// 1. 初始化 $match 条件，如果启用了逻辑租户隔离则注入租户过滤
	matchFilter := bson.M{
		"source_resource_id": id,
		"source_model_uid":   modelUid,
	}
	if tenantID > 0 {
		matchFilter["tenant_id"] = tenantID
	}

	// 2. 初始化 $graphLookup 条件
	graphLookupVal := bson.M{
		"from":             ResourceRelationCollection,
		"startWith":        "$target_resource_id",
		"connectFromField": "target_resource_id",
		"connectToField":   "source_resource_id",
		"as":               "dependencies",
		"maxDepth":         maxDepth,
	}
	if tenantID > 0 {
		graphLookupVal["restrictSearchWithMatch"] = bson.M{"tenant_id": tenantID}
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchFilter}},
		{{Key: "$graphLookup", Value: graphLookupVal}},
	}

	// NOTE: 回归大一统逻辑字段隔离设计，直接在公共集合 Native() 上运行原生聚合，配合 matchFilter 里的租户字段安全过滤
	cursor, err := dao.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("递归下游关系聚合查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	var dbResults []struct {
		ResourceRelation `bson:",inline"`
		Dependencies     []ResourceRelation `bson:"dependencies"`
	}

	if err = cursor.All(ctx, &dbResults); err != nil {
		return nil, fmt.Errorf("递归下游解码错误: %w", err)
	}

	return deduplicateRelations(dbResults), nil
}

func (dao *resourceRelationDAO) ListRecursiveDst(ctx context.Context, modelUid string, id int64, maxDepth int) ([]ResourceRelation, error) {
	tenantID := ctxutil.GetTenantID(ctx).Int64()

	// 1. 初始化 $match 条件，如果启用了逻辑租户隔离则注入租户过滤
	matchFilter := bson.M{
		"target_resource_id": id,
		"target_model_uid":   modelUid,
	}
	if tenantID > 0 {
		matchFilter["tenant_id"] = tenantID
	}

	// 2. 初始化 $graphLookup 条件
	graphLookupVal := bson.M{
		"from":             ResourceRelationCollection,
		"startWith":        "$source_resource_id",
		"connectFromField": "source_resource_id",
		"connectToField":   "target_resource_id",
		"as":               "dependencies",
		"maxDepth":         maxDepth,
	}
	if tenantID > 0 {
		graphLookupVal["restrictSearchWithMatch"] = bson.M{"tenant_id": tenantID}
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchFilter}},
		{{Key: "$graphLookup", Value: graphLookupVal}},
	}

	// NOTE: 回归大一统逻辑字段隔离设计，直接在公共集合 Native() 上运行原生聚合，配合 matchFilter 里的租户字段安全过滤
	cursor, err := dao.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("递归上游关系聚合查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	var dbResults []struct {
		ResourceRelation `bson:",inline"`
		Dependencies     []ResourceRelation `bson:"dependencies"`
	}

	if err = cursor.All(ctx, &dbResults); err != nil {
		return nil, fmt.Errorf("递归上游解码错误: %w", err)
	}

	return deduplicateRelations(dbResults), nil
}

// deduplicateRelations 使用 lo 泛型库优雅实现对包含 Dependencies 的聚合结果展平与唯一去重
func deduplicateRelations(dbResults []struct {
	ResourceRelation `bson:",inline"`
	Dependencies     []ResourceRelation `bson:"dependencies"`
}) []ResourceRelation {
	// 1. 将每一个聚合实体及其下属 dependencies 依赖切片，全部平铺展平为一个统一的切片
	allRelations := lo.FlatMap(dbResults, func(res struct {
		ResourceRelation `bson:",inline"`
		Dependencies     []ResourceRelation `bson:"dependencies"`
	}, _ int) []ResourceRelation {
		return append([]ResourceRelation{res.ResourceRelation}, res.Dependencies...)
	})

	// 2. 利用 lo.UniqBy 依照资源关联 Id 唯一主键进行一键去重，代替繁琐的手写 seen map 循环
	return lo.UniqBy(allRelations, func(item ResourceRelation) int64 {
		return item.Id
	})
}

type ResourceRelation struct {
	TenantID         int64  `bson:"tenant_id"`
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
