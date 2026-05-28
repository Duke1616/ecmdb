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

type RelationTypeDAO interface {
	// Create 创建关联类型
	Create(ctx context.Context, r RelationType) (int64, error)

	// BatchCreate 批量创建
	BatchCreate(ctx context.Context, rts []RelationType) error

	// GetByUids 根据 UID 获取关联类型
	GetByUids(ctx context.Context, uids []string) ([]RelationType, error)

	// List 关联列表
	List(ctx context.Context, offset, limit int64) ([]RelationType, error)

	// Count 数量
	Count(ctx context.Context) (int64, error)

	// Update 更新关联类型
	Update(ctx context.Context, r RelationType) (int64, error)

	// Delete 删除关联类型
	Delete(ctx context.Context, id int64) (int64, error)

	// GetByID 根据 ID 获取关联类型
	GetByID(ctx context.Context, id int64) (RelationType, error)
}

func NewRelationTypeDAO(db *mongox.DB) RelationTypeDAO {
	return &relationDAO{
		db:   db,
		coll: mongox.NewCollection[RelationType](db, RelationTypeCollection),
	}
}

type relationDAO struct {
	db   *mongox.DB
	coll *mongox.Collection[RelationType]
}

func (dao *relationDAO) GetByUids(ctx context.Context, uids []string) ([]RelationType, error) {
	filter := bson.M{
		"uid": bson.M{"$in": uids},
	}
	opts := &options.FindOptions{}

	return dao.coll.Find(ctx, filter, opts)
}

func (dao *relationDAO) BatchCreate(ctx context.Context, rts []RelationType) error {
	if len(rts) == 0 {
		return nil
	}

	now := time.Now().UnixMilli()

	// 批量获取起始 ID
	startID, err := dao.db.GetBatchIdGenerator(RelationTypeCollection, len(rts))
	if err != nil {
		return fmt.Errorf("获取批量 ID 错误: %w", err)
	}

	docs := make([]*RelationType, len(rts))
	for i := range rts {
		rts[i].Id = startID + int64(i)
		rts[i].Ctime, rts[i].Utime = now, now
		docs[i] = &rts[i]
	}

	_, err = dao.coll.InsertMany(ctx, docs)
	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return fmt.Errorf("批量插入关联类型: %w", errs.ErrUniqueDuplicate)
		}
		return fmt.Errorf("批量插入数据错误: %w", err)
	}

	return nil
}

func (dao *relationDAO) Create(ctx context.Context, r RelationType) (int64, error) {
	now := time.Now()
	r.Ctime, r.Utime = now.UnixMilli(), now.UnixMilli()

	// 借助 AutoIDPlugin 插件自动分配自增主键，无须再手动管理自增 ID。
	_, err := dao.coll.InsertOne(ctx, &r)
	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return 0, fmt.Errorf("关联类型插入: %w", errs.ErrUniqueDuplicate)
		}
		return 0, err
	}

	return r.Id, nil
}

func (dao *relationDAO) List(ctx context.Context, offset, limit int64) ([]RelationType, error) {
	filter := bson.M{}
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	return dao.coll.Find(ctx, filter, opts)
}

func (dao *relationDAO) Count(ctx context.Context) (int64, error) {
	filter := bson.M{}
	count, err := dao.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *relationDAO) Update(ctx context.Context, r RelationType) (int64, error) {
	filter := bson.M{"id": r.Id}
	update := bson.M{
		"$set": bson.M{
			"name":            r.Name,
			"source_describe": r.SourceDescribe,
			"target_describe": r.TargetDescribe,
			"utime":           time.Now().UnixMilli(),
		},
	}
	res, err := dao.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("更新关联类型失败: %w", err)
	}
	return res.ModifiedCount, nil
}

func (dao *relationDAO) Delete(ctx context.Context, id int64) (int64, error) {
	filter := bson.M{"id": id}
	res, err := dao.coll.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除关联类型失败: %w", err)
	}
	return res.DeletedCount, nil
}

func (dao *relationDAO) GetByID(ctx context.Context, id int64) (RelationType, error) {
	filter := bson.M{"id": id}
	res, err := dao.coll.FindOne(ctx, filter)
	if err != nil {
		return RelationType{}, fmt.Errorf("查询关联类型失败: %w", err)
	}
	return *res, nil
}

type RelationType struct {
	TenantID       int64  `bson:"tenant_id"`
	Id             int64  `bson:"id"`
	Name           string `bson:"name"`
	UID            string `bson:"uid"`
	SourceDescribe string `bson:"source_describe"`
	TargetDescribe string `bson:"target_describe"`
	Ctime          int64  `bson:"ctime"`
	Utime          int64  `bson:"utime"`
}

func (a *RelationType) SetID(id int64) {
	a.Id = id
}

func (a *RelationType) GetID() int64 {
	return a.Id
}
