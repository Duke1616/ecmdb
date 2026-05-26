package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/ekit/slice"
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

func NewModelGroupDAO(db *mongox.DB) ModelGroupDAO {
	return &groupDAO{
		coll: mongox.NewCollection[ModelGroup](db, ModelGroupCollection),
	}
}

type groupDAO struct {
	coll *mongox.Collection[ModelGroup]
}

func (dao *groupDAO) GetByName(ctx context.Context, name string) (ModelGroup, error) {
	filter := bson.M{"name": name}

	m, err := dao.coll.FindOne(ctx, filter)
	if err != nil {
		if mongox.IsNotFoundError(err) {
			return ModelGroup{}, fmt.Errorf("属性组查询: %w", errs.ErrNotFound)
		}
		return ModelGroup{}, fmt.Errorf("解码错误: %w", err)
	}

	return *m, nil
}

func (dao *groupDAO) BatchCreate(ctx context.Context, mgs []ModelGroup) ([]ModelGroup, error) {
	if len(mgs) == 0 {
		return nil, nil
	}

	now := time.Now().UnixMilli()

	// 依靠 mongoxv2 的 AutoIDPlugin 插件，一次获取批量 ID 并自动设置。
	docs := slice.Map(mgs, func(idx int, src ModelGroup) *ModelGroup {
		return &ModelGroup{
			Name:  mgs[idx].Name,
			Ctime: now,
			Utime: now,
		}
	})

	_, err := dao.coll.InsertMany(ctx, docs)
	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return nil, fmt.Errorf("批量插入属性组: %w", errs.ErrUniqueDuplicate)
		}
		return nil, fmt.Errorf("批量插入数据错误: %w", err)
	}

	result := make([]ModelGroup, len(docs))
	for i, doc := range docs {
		result[i] = *doc
	}

	return result, nil
}

func (dao *groupDAO) GetByNames(ctx context.Context, names []string) ([]ModelGroup, error) {
	filter := bson.M{"name": bson.M{"$in": names}}
	return dao.coll.Find(ctx, filter)
}

func (dao *groupDAO) List(ctx context.Context, offset, limit int64) ([]ModelGroup, error) {
	filter := bson.M{}
	opt := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: 1}},
		Limit: &limit,
		Skip:  &offset,
	}
	return dao.coll.Find(ctx, filter, opt)
}

func (dao *groupDAO) Count(ctx context.Context) (int64, error) {
	filter := bson.M{}

	count, err := dao.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *groupDAO) Delete(ctx context.Context, id int64) (int64, error) {
	filter := bson.M{"id": id}

	result, err := dao.coll.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *groupDAO) CreateModelGroup(ctx context.Context, mg ModelGroup) (int64, error) {
	now := time.Now()
	mg.Ctime, mg.Utime = now.UnixMilli(), now.UnixMilli()

	// 依靠 mongoxv2 的 AutoIDPlugin 插件自动分配并注入 ID，不需要再手动管理 id_generator
	_, err := dao.coll.InsertOne(ctx, &mg)
	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return 0, fmt.Errorf("模型插入: %w", errs.ErrUniqueDuplicate)
		}
		return 0, err
	}

	return mg.Id, nil
}

func (dao *groupDAO) Rename(ctx context.Context, id int64, name string) (int64, error) {
	filter := bson.M{"id": id}
	update := bson.M{
		"$set": bson.M{
			"name":  name,
			"utime": time.Now().UnixMilli(),
		},
	}
	res, err := dao.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("模型组重命名失败: %w", err)
	}
	return res.ModifiedCount, nil
}
