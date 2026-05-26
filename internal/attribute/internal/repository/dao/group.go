package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/samber/lo"
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

	// GetMaxSortKeyByModuleUid 获取分组下的最大 SortKey
	GetMaxSortKeyByModuleUid(ctx context.Context, modelUid string) (int64, error)

	// UpdateSort 更新属性组排序
	UpdateSort(ctx context.Context, id int64, sortKey int64) error

	// BatchUpdateSort 批量更新属性组排序
	BatchUpdateSort(ctx context.Context, items []AttributeGroupSortItem) error
}

type attributeGroupDAO struct {
	db   *mongox.DB
	coll *mongox.Collection[AttributeGroup]
}

func NewAttributeGroupDAO(db *mongox.DB) AttributeGroupDAO {
	return &attributeGroupDAO{
		db:   db,
		coll: mongox.NewCollection[AttributeGroup](db, AttributeGroupCollection),
	}
}

func (dao *attributeGroupDAO) GetMaxSortKeyByModuleUid(ctx context.Context, modelUid string) (int64, error) {
	filter := bson.M{"model_uid": modelUid}
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

func (dao *attributeGroupDAO) BatchCreateAttributeGroup(ctx context.Context, ags []AttributeGroup) ([]AttributeGroup, error) {
	if len(ags) == 0 {
		return nil, nil
	}

	now := time.Now().UnixMilli()

	// 批量获取起始 ID
	startID, err := dao.db.GetBatchIdGenerator(AttributeGroupCollection, len(ags))
	if err != nil {
		return nil, fmt.Errorf("获取批量 ID 错误: %w", err)
	}

	result := make([]AttributeGroup, len(ags))
	docs := make([]*AttributeGroup, len(ags))
	for i := range ags {
		result[i] = ags[i]
		result[i].Id = startID + int64(i)
		result[i].Ctime, result[i].Utime = now, now
		docs[i] = &result[i]
	}

	_, err = dao.coll.InsertMany(ctx, docs)
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

	// 直接插入数据，借助 AutoIDPlugin 插件自动分配自增 ID
	_, err := dao.coll.InsertOne(ctx, &req)
	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return 0, fmt.Errorf("模型插入: %w", errs.ErrUniqueDuplicate)
		}
		return 0, err
	}

	return req.Id, nil
}

func (dao *attributeGroupDAO) ListAttributeGroup(ctx context.Context, modelUid string) ([]AttributeGroup, error) {
	filter := bson.M{"model_uid": modelUid}
	opts := &options.FindOptions{
		Sort: bson.D{{Key: "sort_key", Value: 1}},
	}

	return dao.coll.Find(ctx, filter, opts)
}

func (dao *attributeGroupDAO) ListAttributeGroupByIds(ctx context.Context, ids []int64) ([]AttributeGroup, error) {
	filter := bson.M{"id": bson.M{"$in": ids}}

	return dao.coll.Find(ctx, filter)
}

func (dao *attributeGroupDAO) DeleteAttributeGroup(ctx context.Context, id int64) (int64, error) {
	filter := bson.M{"id": id}

	res, err := dao.coll.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除属性组错误: %w", err)
	}

	return res.DeletedCount, nil
}

func (dao *attributeGroupDAO) RenameAttributeGroup(ctx context.Context, id int64, name string) (int64, error) {
	filter := bson.M{"id": id}
	update := bson.M{
		"$set": bson.M{
			"name":  name,
			"utime": time.Now().UnixMilli(),
		},
	}

	res, err := dao.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("重命名属性组错误: %w", err)
	}

	return res.ModifiedCount, nil
}

func (dao *attributeGroupDAO) UpdateSort(ctx context.Context, id int64, sortKey int64) error {
	filter := bson.M{"id": id}
	update := bson.M{
		"$set": bson.M{
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

func (dao *attributeGroupDAO) BatchUpdateSort(ctx context.Context, items []AttributeGroupSortItem) error {
	if len(items) == 0 {
		return nil
	}

	// NOTE: 使用 lo.Map 代替手动 slice 组装，精简 BulkWrite 的 WriteModel 构造逻辑
	models := lo.Map(items, func(item AttributeGroupSortItem, _ int) mongo.WriteModel {
		return mongo.NewUpdateOneModel().
			SetFilter(bson.M{"id": item.ID}).
			SetUpdate(bson.M{"$set": bson.M{
				"sort_key": item.SortKey,
				"utime":    time.Now().UnixMilli(),
			}})
	})

	_, err := dao.coll.Native().BulkWrite(ctx, models)
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
