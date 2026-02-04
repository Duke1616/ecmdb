package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const AttributeGroupCollection = "c_attribute_group"

type AttributeGroupDAO interface {
	// CreateAttributeGroup 创建属性组
	CreateAttributeGroup(ctx context.Context, req AttributeGroup) (int64, error)

	// BatchCreateAttributeGroup 批量创建组
	BatchCreateAttributeGroup(ctx context.Context, req []AttributeGroup) ([]AttributeGroup, error)

	// ListAttributeGroup 根据模型唯一标识，获取组信息
	ListAttributeGroup(ctx context.Context, modelUid string) ([]AttributeGroup, error)

	// ListAttributeGroupByIds 根据 IDS 获取组信息
	ListAttributeGroupByIds(ctx context.Context, ids []int64) ([]AttributeGroup, error)

	// DeleteAttributeGroup 删除属性组
	DeleteAttributeGroup(ctx context.Context, id int64) (int64, error)

	// RenameAttributeGroup 重命名属性组
	RenameAttributeGroup(ctx context.Context, id int64, name string) (int64, error)

	// UpdateSort 更新属性组排序
	UpdateSort(ctx context.Context, id int64, sortKey int64) error

	// BatchUpdateSort 批量更新属性组排序
	BatchUpdateSort(ctx context.Context, items []AttributeGroupSortItem) error
}

type attributeGroupDAO struct {
	db *mongox.Mongo
}

func NewAttributeGroupDAO(db *mongox.Mongo) AttributeGroupDAO {
	return &attributeGroupDAO{
		db: db,
	}
}

func (dao *attributeGroupDAO) BatchCreateAttributeGroup(ctx context.Context, ags []AttributeGroup) ([]AttributeGroup, error) {
	if len(ags) == 0 {
		return nil, nil
	}

	col := dao.db.Collection(AttributeGroupCollection)
	now := time.Now().UnixMilli()

	// 批量获取起始 ID（一次数据库调用）
	startID, err := dao.db.GetBatchIdGenerator(AttributeGroupCollection, len(ags))
	if err != nil {
		return nil, fmt.Errorf("获取批量 ID 错误: %w", err)
	}

	// 为每个属性设置 ID 和时间戳
	result := make([]AttributeGroup, len(ags))
	docs := make([]interface{}, len(ags))
	for i := range ags {
		// 创建新的对象，避免修改原参数
		result[i] = ags[i]
		result[i].Id = startID + int64(i)
		result[i].Ctime, result[i].Utime = now, now
		docs[i] = result[i]
	}

	// 执行批量插入
	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return nil, fmt.Errorf("批量插入属性组: %w", errs.ErrUniqueDuplicate)
		}
		return nil, fmt.Errorf("批量插入数据错误: %w", err)
	}

	return result, nil
}

func (dao *attributeGroupDAO) CreateAttributeGroup(ctx context.Context, req AttributeGroup) (int64, error) {
	now := time.Now()
	req.Ctime, req.Utime = now.UnixMilli(), now.UnixMilli()

	// 直接插入数据，并自增ID
	_, err := dao.db.InsertOneWithAutoID(ctx, AttributeGroupCollection, &req)

	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return 0, fmt.Errorf("模型插入: %w", errs.ErrUniqueDuplicate)
		}
		return 0, err
	}

	return req.Id, nil
}

func (dao *attributeGroupDAO) ListAttributeGroup(ctx context.Context, modelUid string) ([]AttributeGroup, error) {
	col := dao.db.Collection(AttributeGroupCollection)
	filter := bson.M{"model_uid": modelUid}
	opts := &options.FindOptions{
		Sort: bson.D{{Key: "sort_key", Value: 1}},
	}

	cursor, err := col.Find(ctx, filter, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []AttributeGroup
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *attributeGroupDAO) ListAttributeGroupByIds(ctx context.Context, ids []int64) ([]AttributeGroup, error) {
	col := dao.db.Collection(AttributeGroupCollection)
	filter := bson.M{"id": bson.M{"$in": ids}}

	cursor, err := col.Find(ctx, filter)
	var result []AttributeGroup
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *attributeGroupDAO) DeleteAttributeGroup(ctx context.Context, id int64) (int64, error) {
	col := dao.db.Collection(AttributeGroupCollection)
	filter := bson.M{"id": id}

	res, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除属性组错误: %w", err)
	}

	return res.DeletedCount, nil
}

func (dao *attributeGroupDAO) RenameAttributeGroup(ctx context.Context, id int64, name string) (int64, error) {
	col := dao.db.Collection(AttributeGroupCollection)
	filter := bson.M{"id": id}
	update := bson.M{
		"$set": bson.M{
			"name":  name,
			"utime": time.Now().UnixMilli(),
		},
	}

	res, err := col.UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("重命名属性组错误: %w", err)
	}

	return res.ModifiedCount, nil
}

func (dao *attributeGroupDAO) UpdateSort(ctx context.Context, id int64, sortKey int64) error {
	col := dao.db.Collection(AttributeGroupCollection)
	filter := bson.M{"id": id}
	update := bson.M{
		"$set": bson.M{
			"sort_key": sortKey,
			"utime":    time.Now().UnixMilli(),
		},
	}

	_, err := col.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("更新排序错误: %w", err)
	}
	return nil
}

func (dao *attributeGroupDAO) BatchUpdateSort(ctx context.Context, items []AttributeGroupSortItem) error {
	if len(items) == 0 {
		return nil
	}

	col := dao.db.Collection(AttributeGroupCollection)
	var models []mongo.WriteModel
	for _, item := range items {
		update := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"id": item.ID}).
			SetUpdate(bson.M{"$set": bson.M{
				"sort_key": item.SortKey,
				"utime":    time.Now().UnixMilli(),
			}})
		models = append(models, update)
	}

	_, err := col.BulkWrite(ctx, models)
	if err != nil {
		return fmt.Errorf("批量更新排序错误: %w", err)
	}
	return nil
}

type AttributeGroup struct {
	Id       int64  `bson:"id"`
	Name     string `bson:"name"`
	ModelUid string `bson:"model_uid"`
	SortKey  int64  `bson:"sort_key"`
	Ctime    int64  `bson:"ctime"`
	Utime    int64  `bson:"utime"`
}

func (a *AttributeGroup) SetID(id int64) {
	a.Id = id
}

func (a *AttributeGroup) GetID() int64 {
	return a.Id
}

// AttributeGroupSortItem 排序更新项
type AttributeGroupSortItem struct {
	ID      int64
	SortKey int64
}
