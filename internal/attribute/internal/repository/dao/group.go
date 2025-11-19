package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
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

	col := dao.db.Collection(AttributeCollection)
	now := time.Now().UnixMilli()

	// 批量获取起始 ID（一次数据库调用）
	startID, err := dao.db.GetBatchIdGenerator(AttributeCollection, len(ags))
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
	opts := &options.FindOptions{}

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

type AttributeGroup struct {
	Id       int64  `bson:"id"`
	Name     string `bson:"name"`
	ModelUid string `bson:"model_uid"`
	Index    int64  `bson:"index"`
	Ctime    int64  `bson:"ctime"`
	Utime    int64  `bson:"utime"`
}

func (a *AttributeGroup) SetID(id int64) {
	a.Id = id
}

func (a *AttributeGroup) GetID() int64 {
	return a.Id
}
