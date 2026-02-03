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

type ModelGroupDAO interface {
	// CreateModelGroup 创建模型分组
	CreateModelGroup(ctx context.Context, mg ModelGroup) (int64, error)

	// BatchCreate 批量创建模型分组
	BatchCreate(ctx context.Context, req []ModelGroup) ([]ModelGroup, error)

	// GetByNames 根据名称查询模型组
	GetByNames(ctx context.Context, names []string) ([]ModelGroup, error)

	// GetByName 根据名称查询模型组
	GetByName(ctx context.Context, name string) (ModelGroup, error)

	// List 模型列表，支持分页
	List(ctx context.Context, offset, limit int64) ([]ModelGroup, error)

	// Count 模型分组数量
	Count(ctx context.Context) (int64, error)

	// Delete 删除分组
	Delete(ctx context.Context, id int64) (int64, error)

	// Rename 重命名分组
	Rename(ctx context.Context, id int64, name string) (int64, error)
}

func NewModelGroupDAO(db *mongox.Mongo) ModelGroupDAO {
	return &groupDAO{
		db: db,
	}
}

type groupDAO struct {
	db *mongox.Mongo
}

func (dao *groupDAO) GetByName(ctx context.Context, name string) (ModelGroup, error) {
	col := dao.db.Collection(ModelGroupCollection)
	filter := bson.M{"name": name}

	var result ModelGroup
	if err := col.FindOne(ctx, filter).Decode(&result); err != nil {
		if mongox.IsNotFoundError(err) {
			return ModelGroup{}, fmt.Errorf("属性组查询: %w", errs.ErrNotFound)
		}
		return ModelGroup{}, fmt.Errorf("解码错误: %w", err)
	}

	return result, nil
}

func (dao *groupDAO) BatchCreate(ctx context.Context, mgs []ModelGroup) ([]ModelGroup, error) {
	if len(mgs) == 0 {
		return nil, nil
	}

	col := dao.db.Collection(ModelGroupCollection)
	now := time.Now().UnixMilli()

	// 批量获取起始 ID（一次数据库调用）
	startID, err := dao.db.GetBatchIdGenerator(ModelGroupCollection, len(mgs))
	if err != nil {
		return nil, fmt.Errorf("获取批量 ID 错误: %w", err)
	}

	// 为每个属性设置 ID 和时间戳
	result := make([]ModelGroup, len(mgs))
	docs := make([]interface{}, len(mgs))
	for i := range mgs {
		// 创建新的对象，避免修改原参数
		result[i] = mgs[i]
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

func (dao *groupDAO) GetByNames(ctx context.Context, names []string) ([]ModelGroup, error) {
	col := dao.db.Collection(ModelGroupCollection)
	filter := bson.M{"name": bson.M{"$in": names}}
	cursor, err := col.Find(ctx, filter)
	var result []ModelGroup
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *groupDAO) List(ctx context.Context, offset, limit int64) ([]ModelGroup, error) {
	col := dao.db.Collection(ModelGroupCollection)
	filer := bson.M{}
	opt := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: 1}},
		Limit: &limit,
		Skip:  &offset,
	}
	cursor, err := col.Find(ctx, filer, opt)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []ModelGroup
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *groupDAO) Count(ctx context.Context) (int64, error) {
	col := dao.db.Collection(ModelGroupCollection)
	filter := bson.M{}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *groupDAO) Delete(ctx context.Context, id int64) (int64, error) {
	col := dao.db.Collection(ModelGroupCollection)
	filter := bson.M{"id": id}

	result, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *groupDAO) CreateModelGroup(ctx context.Context, mg ModelGroup) (int64, error) {
	now := time.Now()
	mg.Ctime, mg.Utime = now.UnixMilli(), now.UnixMilli()

	// 直接插入数据，并自增ID
	_, err := dao.db.InsertOneWithAutoID(ctx, ModelGroupCollection, &mg)

	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return 0, fmt.Errorf("模型插入: %w", errs.ErrUniqueDuplicate)
		}
		return 0, err
	}

	return mg.Id, nil
}

func (dao *groupDAO) Rename(ctx context.Context, id int64, name string) (int64, error) {
	col := dao.db.Collection(ModelGroupCollection)
	filter := bson.M{"id": id}
	update := bson.M{
		"$set": bson.M{
			"name":  name,
			"utime": time.Now().UnixMilli(),
		},
	}
	res, err := col.UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("模型组重命名失败: %w", err)
	}
	return res.ModifiedCount, nil
}
