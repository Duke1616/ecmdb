package dao

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ModelDAO interface {
	// Create 创建模型
	Create(ctx context.Context, m Model) (int64, error)

	// GetById 根据 ID 获取模型详情
	GetById(ctx context.Context, id int64) (Model, error)

	// GetByUids 根据模型唯一标识组，获取模型组
	GetByUids(ctx context.Context, uids []string) ([]Model, error)

	// GetByUid 根据唯一标识获取模型详情
	GetByUid(ctx context.Context, uid string) (Model, error)

	// List 获取模型列表，支持分页
	List(ctx context.Context, offset, limit int64) ([]Model, error)

	// Count 获取模型数量
	Count(ctx context.Context) (int64, error)

	// ListAll 获取所有模型，不分页
	ListAll(ctx context.Context) ([]Model, error)

	// ListByGroupIds 获取指定组下的模型
	ListByGroupIds(ctx context.Context, mgids []int64) ([]Model, error)

	// DeleteById 根据ID删除模型
	DeleteById(ctx context.Context, id int64) (int64, error)

	// DeleteByUid 根据唯一标识删除模型
	DeleteByUid(ctx context.Context, modelUid string) (int64, error)

	// CountByGroupId 获取指定组下的模型数量
	CountByGroupId(ctx context.Context, GroupId int64) (int64, error)
}

func NewModelDAO(db *mongox.DB) ModelDAO {
	return &modelDAO{
		coll: mongox.NewCollection[Model](db, ModelCollection),
	}
}

type modelDAO struct {
	coll *mongox.Collection[Model]
}

func (dao *modelDAO) GetByUid(ctx context.Context, uid string) (Model, error) {
	filter := bson.M{"uid": uid}

	m, err := dao.coll.FindOne(ctx, filter)
	if err != nil {
		if mongox.IsNotFoundError(err) {
			return Model{}, fmt.Errorf("模型查询: %w", errs.ErrNotFound)
		}
		return Model{}, fmt.Errorf("解码错误: %w", err)
	}

	return *m, nil
}

func (dao *modelDAO) GetByUids(ctx context.Context, uids []string) ([]Model, error) {
	filter := bson.M{"uid": bson.M{"$in": uids}}
	ms, err := dao.coll.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("查询错误: %w", err)
	}
	return ms, nil
}

func (dao *modelDAO) ListByGroupIds(ctx context.Context, mgids []int64) ([]Model, error) {
	if len(mgids) <= 0 {
		slog.Warn("没有匹配的数据, 模型组为空")
		return nil, nil
	}

	filter := bson.M{
		"model_group_id": bson.M{
			"$in": mgids,
		},
	}

	opts := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: 1}},
	}

	ms, err := dao.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误: %w", err)
	}
	return ms, nil
}

func (dao *modelDAO) Create(ctx context.Context, m Model) (int64, error) {
	now := time.Now()
	m.Ctime, m.Utime = now.UnixMilli(), now.UnixMilli()

	// 依靠 mongoxv2 的 AutoIDPlugin 插件自动分配并注入 ID，不需要再手动管理 id_generator
	_, err := dao.coll.InsertOne(ctx, &m)
	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return 0, fmt.Errorf("模型插入: %w", errs.ErrUniqueDuplicate)
		}
		return 0, err
	}

	return m.Id, nil
}

func (dao *modelDAO) GetById(ctx context.Context, id int64) (Model, error) {
	filter := bson.M{"id": id}

	m, err := dao.coll.FindOne(ctx, filter)
	if err != nil {
		if mongox.IsNotFoundError(err) {
			return Model{}, fmt.Errorf("模型查询: %w", errs.ErrNotFound)
		}
		return Model{}, fmt.Errorf("解码错误: %w", err)
	}

	return *m, nil
}

func (dao *modelDAO) ListAll(ctx context.Context) ([]Model, error) {
	filter := bson.M{}
	opt := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}
	ms, err := dao.coll.Find(ctx, filter, opt)
	if err != nil {
		return nil, fmt.Errorf("查询错误: %w", err)
	}
	return ms, nil
}

func (dao *modelDAO) List(ctx context.Context, offset, limit int64) ([]Model, error) {
	filter := bson.M{}
	opt := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}
	ms, err := dao.coll.Find(ctx, filter, opt)
	if err != nil {
		return nil, fmt.Errorf("查询错误: %w", err)
	}
	return ms, nil
}

func (dao *modelDAO) Count(ctx context.Context) (int64, error) {
	filter := bson.M{}

	count, err := dao.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *modelDAO) DeleteById(ctx context.Context, id int64) (int64, error) {
	filter := bson.M{"id": id}

	result, err := dao.coll.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *modelDAO) DeleteByUid(ctx context.Context, modelUid string) (int64, error) {
	filter := bson.M{"uid": modelUid}

	result, err := dao.coll.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *modelDAO) CountByGroupId(ctx context.Context, GroupId int64) (int64, error) {
	filter := bson.M{"model_group_id": GroupId}
	count, err := dao.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}
	return count, nil
}
