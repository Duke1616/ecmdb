package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const AttributeCollection = "c_attribute"

type AttributeDAO interface {
	// CreateAttribute 创建模型属性字段
	CreateAttribute(ctx context.Context, ab Attribute) (int64, error)

	// BatchCreateAttribute 批量创建模型属性字段
	BatchCreateAttribute(ctx context.Context, attrs []Attribute) error

	// SearchAttributeByModelUID 根据模型唯一值，排除安全字段
	SearchAttributeByModelUID(ctx context.Context, modelUid string) ([]Attribute, error)

	// SearchAttributeFieldsBySecure 根据模型唯一值，仅展示安全字段
	SearchAttributeFieldsBySecure(ctx context.Context, modelUids []string) ([]Attribute, error)

	// ListAttributes 根据模型唯一值，搜索所有字段
	ListAttributes(ctx context.Context, modelUid string) ([]Attribute, error)

	// Count 根据模型唯一值，查看字段数量
	Count(ctx context.Context, modelUid string) (int64, error)

	// UpdateFieldIndex 自定义展示字段
	UpdateFieldIndex(ctx context.Context, modelUid string, customField []string) (int64, error)

	// UpdateFieldIndexReverse 变更顺序
	UpdateFieldIndexReverse(ctx context.Context, modelUid string, customField []string) (int64, error)

	// DeleteAttribute 根据 ID 删除模型字段
	DeleteAttribute(ctx context.Context, id int64) (int64, error)

	// ListAttributePipeline 根据模型字段分组，进行聚合返回
	ListAttributePipeline(ctx context.Context, modelUid string) ([]AttributePipeline, error)

	// UpdateAttribute 修改属性
	UpdateAttribute(ctx context.Context, attribute Attribute) (int64, error)

	// DetailAttribute 根据 ID 查看详情属性
	DetailAttribute(ctx context.Context, id int64) (Attribute, error)

	// DeleteByGroupId 根据分组ID删除所有属性
	DeleteByGroupId(ctx context.Context, groupId int64) (int64, error)

	// ListByGroupID 根据分组ID获取属性列表（按 SortKey 排序）
	ListByGroupID(ctx context.Context, groupId int64) ([]Attribute, error)

	// GetMaxSortKeyByGroupID 获取分组下的最大 SortKey
	GetMaxSortKeyByGroupID(ctx context.Context, groupId int64) (int64, error)

	// UpdateSort 更新属性的分组和排序
	UpdateSort(ctx context.Context, id, groupId, sortKey int64) error

	// BatchUpdateSortKey 批量更新属性的 SortKey
	BatchUpdateSortKey(ctx context.Context, items []AttributeSortItem) error
}

var ErrVersionConflict = errors.New("attribute version conflict")

type attributeDAO struct {
	db   *mongox.DB
	coll *mongox.Collection[Attribute]
}

func NewAttributeDAO(db *mongox.DB) AttributeDAO {
	return &attributeDAO{
		db:   db,
		coll: mongox.NewCollection[Attribute](db, AttributeCollection),
	}
}

func (dao *attributeDAO) DetailAttribute(ctx context.Context, id int64) (Attribute, error) {
	filter := bson.M{"id": id}

	result, err := dao.coll.FindOne(ctx, filter)
	if err != nil {
		if mongox.IsNotFoundError(err) {
			return Attribute{}, fmt.Errorf("属性查询: %w", errs.ErrNotFound)
		}
		return Attribute{}, fmt.Errorf("解码错误: %w", err)
	}

	return *result, nil
}

func (dao *attributeDAO) UpdateAttribute(ctx context.Context, attr Attribute) (int64, error) {
	now := time.Now().UnixMilli()
	filter := bson.M{
		"id": attr.Id,
	}

	// 新数据走 CAS
	if attr.Version > 0 {
		filter["version"] = attr.Version
	} else {
		// 兼容历史数据
		filter["$or"] = []bson.M{
			{"version": bson.M{"$exists": false}},
			{"version": 0},
		}
	}

	updateDoc := bson.M{
		"$set": bson.M{
			"field_name": attr.FieldName,
			"field_type": attr.FieldType,
			"required":   attr.Required,
			"secure":     attr.Secure,
			"link":       attr.Link,
			"option":     attr.Option,
			"utime":      now,
		},
		"$inc": bson.M{
			"version": 1,
		},
	}

	res, err := dao.coll.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作失败: %w", err)
	}

	if res.MatchedCount == 0 {
		return 0, ErrVersionConflict
	}

	return res.ModifiedCount, nil
}

func (dao *attributeDAO) CreateAttribute(ctx context.Context, attr Attribute) (int64, error) {
	now := time.Now()
	attr.Ctime, attr.Utime = now.UnixMilli(), now.UnixMilli()

	// 借助 AutoIDPlugin 插件自动分配自增主键，无须再手动管理自增 ID。
	_, err := dao.coll.InsertOne(ctx, &attr)
	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return 0, fmt.Errorf("模型插入: %w", errs.ErrUniqueDuplicate)
		}
		return 0, err
	}

	return attr.Id, nil
}

func (dao *attributeDAO) BatchCreateAttribute(ctx context.Context, attrs []Attribute) error {
	if len(attrs) == 0 {
		return nil
	}

	now := time.Now().UnixMilli()

	docs := make([]*Attribute, len(attrs))
	for i := range attrs {
		attrs[i].Ctime, attrs[i].Utime = now, now
		docs[i] = &attrs[i]
	}

	_, err := dao.coll.InsertMany(ctx, docs)
	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return fmt.Errorf("批量插入模型属性: %w", errs.ErrUniqueDuplicate)
		}
		return fmt.Errorf("批量插入数据错误: %w", err)
	}

	return nil
}

func (dao *attributeDAO) SearchAttributeByModelUID(ctx context.Context, modelUid string) ([]Attribute, error) {
	filter := bson.M{
		"model_uid": modelUid,
		"secure":    bson.M{"$ne": true},
	}
	opt := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	return dao.coll.Find(ctx, filter, opt)
}

func (dao *attributeDAO) ListAttributes(ctx context.Context, modelUid string) ([]Attribute, error) {
	filter := bson.M{"model_uid": modelUid}
	opts := &options.FindOptions{
		Sort: bson.D{{Key: "sort_key", Value: 1}},
	}

	return dao.coll.Find(ctx, filter, opts)
}

func (dao *attributeDAO) Count(ctx context.Context, modelUid string) (int64, error) {
	filter := bson.M{"model_uid": modelUid}

	count, err := dao.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *attributeDAO) UpdateFieldIndex(ctx context.Context, modelUid string, customField []string) (int64, error) {
	// NOTE: 使用 lo.Map 精简 BulkWrite 的 WriteModel 构造逻辑
	updates := lo.Map(customField, func(src string, idx int) mongo.WriteModel {
		return &mongo.UpdateOneModel{
			Filter: bson.D{{Key: "model_uid", Value: modelUid}, {Key: "field_name", Value: src}},
			Update: bson.D{{Key: "$set", Value: bson.D{{Key: "display", Value: true}, {Key: "index", Value: idx}}}},
		}
	})

	result, err := dao.coll.BulkWrite(ctx, updates)
	if err != nil {
		return 0, fmt.Errorf("BulkWrite 修改错误, %w", err)
	}

	return result.ModifiedCount, nil
}

func (dao *attributeDAO) UpdateFieldIndexReverse(ctx context.Context, modelUid string, customField []string) (int64, error) {
	filter := bson.D{
		{Key: "model_uid", Value: modelUid},
		{Key: "field_name", Value: bson.D{{Key: "$nin", Value: customField}}},
	}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "display", Value: false}}}}

	// 执行批量更新
	result, err := dao.coll.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("取反修改错误, %w", err)
	}

	return result.ModifiedCount, nil
}

func (dao *attributeDAO) DeleteAttribute(ctx context.Context, id int64) (int64, error) {
	filter := bson.M{"id": id}

	result, err := dao.coll.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *attributeDAO) DeleteByGroupId(ctx context.Context, groupId int64) (int64, error) {
	filter := bson.M{"group_id": groupId}

	result, err := dao.coll.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("批量删除属性失败: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *attributeDAO) ListAttributePipeline(ctx context.Context, modelUid string) ([]AttributePipeline, error) {
	filter := bson.D{
		{Key: "model_uid", Value: modelUid},
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$sort", Value: bson.D{{Key: "sort_key", Value: 1}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "$toLong", Value: "$group_id"}}},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "attributes", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
		}}},
	}

	cursor, err := dao.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}
	defer cursor.Close(ctx)

	var result []AttributePipeline
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}

	return result, nil
}

func (dao *attributeDAO) SearchAttributeFieldsBySecure(ctx context.Context, modelUids []string) ([]Attribute, error) {
	filter := bson.M{
		"secure":    bson.M{"$eq": true},
		"model_uid": bson.M{"$in": modelUids},
	}

	opt := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	return dao.coll.Find(ctx, filter, opt)
}

type Attribute struct {
	TenantID  int64       `bson:"tenant_id"`
	Id        int64       `bson:"id"`
	GroupId   int64       `bson:"group_id"`   // 属性所属分组
	ModelUID  string      `bson:"model_uid"`  // 模型唯一标识
	FieldUid  string      `bson:"field_uid"`  // 字段唯一标识
	FieldName string      `bson:"field_name"` // 字段名称
	FieldType string      `bson:"field_type"` // 字段类型
	Required  bool        `bson:"required"`   // 是否为必传
	Display   bool        `bson:"display"`    // 是否前端展示
	Index     int64       `bson:"index"`      // 字段前端展示顺序
	SortKey   int64       `bson:"sort_key"`   // 拖拽排序键（稀疏索引）
	Secure    bool        `bson:"secure"`     // 是否字段安全、脱敏、加密
	Link      bool        `bson:"link"`       // 是否外链
	Builtin   bool        `bson:"builtin"`    // 是否内置属性
	Version   int64       `bson:"version"`    // CAS 操作
	Option    interface{} `bson:"option"`     // TODO: 为了后续扩展，不同类型的 option 可能不同
	Ctime     int64       `bson:"ctime"`
	Utime     int64       `bson:"utime"`
}

type AttributePipeline struct {
	GroupId    int64       `bson:"_id"`
	Total      int         `bson:"total"`
	Attributes []Attribute `bson:"attributes"`
}

// AttributeSortItem 排序更新项
type AttributeSortItem struct {
	ID      int64
	GroupId int64
	SortKey int64
}

func (a *Attribute) SetID(id int64) {
	a.Id = id
}

func (a *Attribute) GetID() int64 {
	return a.Id
}

func (dao *attributeDAO) ListByGroupID(ctx context.Context, groupId int64) ([]Attribute, error) {
	filter := bson.M{"group_id": groupId}
	opts := &options.FindOptions{
		Sort: bson.D{{Key: "sort_key", Value: 1}}, // 按 sort_key 升序
	}

	return dao.coll.Find(ctx, filter, opts)
}

func (dao *attributeDAO) GetMaxSortKeyByGroupID(ctx context.Context, groupId int64) (int64, error) {
	filter := bson.M{"group_id": groupId}
	opts := &options.FindOneOptions{
		Sort: bson.D{{Key: "sort_key", Value: -1}}, // 按 sort_key 降序取第一个
	}

	result, err := dao.coll.FindOne(ctx, filter, opts)
	if err != nil {
		if mongox.IsNotFoundError(err) {
			return 0, nil // 分组为空时返回 0
		}
		return 0, fmt.Errorf("查询最大 SortKey 错误: %w", err)
	}
	return result.SortKey, nil
}

func (dao *attributeDAO) UpdateSort(ctx context.Context, id, groupId, sortKey int64) error {
	filter := bson.M{"id": id}
	update := bson.M{
		"$set": bson.M{
			"group_id": groupId,
			"sort_key": sortKey,
			"utime":    time.Now().UnixMilli(),
		},
	}

	_, err := dao.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("更新排序错误: %w", err)
	}
	return nil
}

func (dao *attributeDAO) BatchUpdateSortKey(ctx context.Context, items []AttributeSortItem) error {
	if len(items) == 0 {
		return nil
	}

	// 借助 lo.Map 优雅构造 BulkWrite 序列化模型
	models := lo.Map(items, func(item AttributeSortItem, _ int) mongo.WriteModel {
		return mongo.NewUpdateOneModel().
			SetFilter(bson.M{"id": item.ID}).
			SetUpdate(bson.M{"$set": bson.M{
				"group_id": item.GroupId,
				"sort_key": item.SortKey,
				"utime":    time.Now().UnixMilli(),
			}})
	})

	_, err := dao.coll.BulkWrite(ctx, models)
	if err != nil {
		return fmt.Errorf("批量更新 SortKey 错误: %w", err)
	}
	return nil
}
